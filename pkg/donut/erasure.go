package donut

import (
	"bytes"
	"io"
	"strconv"
	"time"

	"github.com/minio-io/mc/pkg/encoding/erasure"
	"github.com/minio-io/mc/pkg/utils/split"
)

func erasureReader(readers []io.ReadCloser, donutMetadata map[string]string, writer *io.PipeWriter) {
	totalChunks, _ := strconv.Atoi(donutMetadata["chunkCount"])
	totalLeft, _ := strconv.Atoi(donutMetadata["totalLength"])
	blockSize, _ := strconv.Atoi(donutMetadata["blockSize"])
	params, _ := erasure.ParseEncoderParams(8, 8, erasure.Cauchy)
	encoder := erasure.NewEncoder(params)
	for _, reader := range readers {
		defer reader.Close()
	}
	for i := 0; i < totalChunks; i++ {
		encodedBytes := make([][]byte, 16)
		for i, reader := range readers {
			var bytesBuffer bytes.Buffer
			io.Copy(&bytesBuffer, reader)
			encodedBytes[i] = bytesBuffer.Bytes()
		}
		curBlockSize := totalLeft
		if blockSize < totalLeft {
			curBlockSize = blockSize
		}
		decodedData, err := encoder.Decode(encodedBytes, curBlockSize)
		if err != nil {
			writer.CloseWithError(err)
			return
		}
		io.Copy(writer, bytes.NewBuffer(decodedData))
		totalLeft = totalLeft - blockSize
	}
	writer.Close()
}

// erasure writer

type erasureWriter struct {
	writers       []Writer
	metadata      map[string]string
	donutMetadata map[string]string // not exposed
	writer        *io.PipeWriter
	isClosed      <-chan bool
}

func newErasureWriter(writers []Writer) ObjectWriter {
	r, w := io.Pipe()
	isClosed := make(chan bool)
	writer := erasureWriter{
		writers:  writers,
		metadata: make(map[string]string),
		writer:   w,
		isClosed: isClosed,
	}
	go erasureGoroutine(r, writer, isClosed)
	return writer
}

func erasureGoroutine(r *io.PipeReader, eWriter erasureWriter, isClosed chan<- bool) {
	chunks := split.Stream(r, 10*1024*1024)
	params, _ := erasure.ParseEncoderParams(8, 8, erasure.Cauchy)
	encoder := erasure.NewEncoder(params)
	chunkCount := 0
	totalLength := 0
	for chunk := range chunks {
		if chunk.Err == nil {
			totalLength = totalLength + len(chunk.Data)
			encodedBlocks, _ := encoder.Encode(chunk.Data)
			for blockIndex, block := range encodedBlocks {
				io.Copy(eWriter.writers[blockIndex], bytes.NewBuffer(block))
			}
		}
		chunkCount = chunkCount + 1
	}
	metadata := make(map[string]string)
	metadata["blockSize"] = strconv.Itoa(10 * 1024 * 1024)
	metadata["chunkCount"] = strconv.Itoa(chunkCount)
	metadata["created"] = time.Now().Format(time.RFC3339Nano)
	metadata["erasureK"] = "8"
	metadata["erasureM"] = "8"
	metadata["erasureTechnique"] = "Cauchy"
	metadata["totalLength"] = strconv.Itoa(totalLength)
	for _, nodeWriter := range eWriter.writers {
		if nodeWriter != nil {
			nodeWriter.SetMetadata(eWriter.metadata)
			nodeWriter.SetDonutDriverMetadata(metadata)
			nodeWriter.Close()
		}
	}
	isClosed <- true
}

func (d erasureWriter) Write(data []byte) (int, error) {
	io.Copy(d.writer, bytes.NewBuffer(data))
	return len(data), nil
}

func (d erasureWriter) Close() error {
	d.writer.Close()
	<-d.isClosed
	return nil
}

func (d erasureWriter) CloseWithError(err error) error {
	for _, writer := range d.writers {
		if writer != nil {
			writer.CloseWithError(err)
		}
	}
	return nil
}

func (d erasureWriter) SetMetadata(metadata map[string]string) error {
	for k := range d.metadata {
		delete(d.metadata, k)
	}
	for k, v := range metadata {
		d.metadata[k] = v
	}
	return nil
}

func (d erasureWriter) GetMetadata() (map[string]string, error) {
	metadata := make(map[string]string)
	for k, v := range d.metadata {
		metadata[k] = v
	}
	return metadata, nil
}
