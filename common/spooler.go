package common

import (
	"io"

	spoolerpkg "github.com/tliron/kubernetes-registry-spooler/client"
	"github.com/tliron/puccini/common/registry"
	urlpkg "github.com/tliron/puccini/url"
)

func PushToRegistry(imageName string, url urlpkg.URL, spooler *spoolerpkg.Client) error {
	reader, err := url.Open()
	if err != nil {
		return err
	}

	reader, err = url.Open()
	if err != nil {
		return err
	}

	if readCloser, ok := reader.(io.ReadCloser); ok {
		defer readCloser.Close()
	}

	if err = spooler.Push(imageName, reader); err == nil {
		return nil
	} else {
		return err
	}
}

func PullLayerFromRegistry(imageName string, writer io.Writer, spooler *spoolerpkg.Client) error {
	pipeReader, pipeWriter := io.Pipe()

	go func() {
		if err := spooler.PullTarball(imageName, pipeWriter); err != nil {
			pipeWriter.Close()
		} else {
			pipeWriter.CloseWithError(err)
		}
	}()

	decoder := registry.NewImageLayerDecoder(pipeReader)
	if _, err := io.Copy(writer, decoder.Decode()); err == nil {
		return nil
	} else {
		return err
	}
}

// TODO: unused. unnecessary?
func TarAndPushToRegistry(imageName string, url urlpkg.URL, spooler *spoolerpkg.Client) error {
	reader, err := url.Open()
	if err != nil {
		return err
	}

	size, err := ReaderSize(reader)
	if err != nil {
		return err
	}

	if readCloser, ok := reader.(io.ReadCloser); ok {
		if err := readCloser.Close(); err != nil {
			return err
		}
	}

	reader, err = url.Open()
	if err != nil {
		return err
	}

	if readCloser, ok := reader.(io.ReadCloser); ok {
		defer readCloser.Close()
	}

	encoder := NewTarEncoder(reader, size)
	if err = spooler.Push(imageName, encoder.Encode()); err == nil {
		return nil
	} else {
		return err
	}
}
