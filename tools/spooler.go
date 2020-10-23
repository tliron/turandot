package tools

import (
	"io"

	spoolerpkg "github.com/tliron/kubernetes-registry-spooler/client"
	urlpkg "github.com/tliron/kutil/url"
	"github.com/tliron/kutil/util"
)

func PublishOnRegistry(imageName string, url urlpkg.URL, spooler *spoolerpkg.Client) error {
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

	if err = spooler.Publish(imageName, reader); err == nil {
		return nil
	} else {
		return err
	}
}

/*
func PullLayerFromRegistry(imageName string, writer io.Writer, spooler *spoolerpkg.Client) error {
	pipeReader, pipeWriter := io.Pipe()

	go func() {
		if err := spooler.PullTarball(imageName, pipeWriter); err != nil {
			pipeWriter.Close()
		} else {
			pipeWriter.CloseWithError(err)
		}
	}()

	decoder := urlpkg.NewContainerImageLayerDecoder(pipeReader)
	if _, err := io.Copy(writer, decoder.Decode()); err == nil {
		return nil
	} else {
		return err
	}
}
*/

// TODO: unused. unnecessary?
func TarAndPublishOnRegistry(imageName string, url urlpkg.URL, spooler *spoolerpkg.Client) error {
	reader, err := url.Open()
	if err != nil {
		return err
	}

	size, err := util.ReaderSize(reader)
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

	encoder := util.NewTarEncoder(reader, "portable", size)
	if err = spooler.Publish(imageName, encoder.Encode()); err == nil {
		return nil
	} else {
		return err
	}
}
