package ssh

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
)

// UploadFile copies a local file to a remote path via SFTP.
func (c *Client) UploadFile(localPath, remotePath string) error {
	sftpClient, err := sftp.NewClient(c.conn)
	if err != nil {
		return fmt.Errorf("sftp client: %w", err)
	}
	defer sftpClient.Close()

	// Ensure remote directory exists.
	remoteDir := filepath.Dir(remotePath)
	if err := sftpClient.MkdirAll(remoteDir); err != nil {
		return fmt.Errorf("mkdir %s: %w", remoteDir, err)
	}

	src, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open local file %s: %w", localPath, err)
	}
	defer src.Close()

	dst, err := sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("create remote file %s: %w", remotePath, err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("copy to %s: %w", remotePath, err)
	}
	return nil
}

// UploadBytes writes a byte slice to a remote path via SFTP.
func (c *Client) UploadBytes(data []byte, remotePath string) error {
	sftpClient, err := sftp.NewClient(c.conn)
	if err != nil {
		return fmt.Errorf("sftp client: %w", err)
	}
	defer sftpClient.Close()

	remoteDir := filepath.Dir(remotePath)
	if err := sftpClient.MkdirAll(remoteDir); err != nil {
		return fmt.Errorf("mkdir %s: %w", remoteDir, err)
	}

	dst, err := sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("create remote file %s: %w", remotePath, err)
	}
	defer dst.Close()

	if _, err := dst.Write(data); err != nil {
		return fmt.Errorf("write to %s: %w", remotePath, err)
	}
	return nil
}
