package image

import _ "embed"

//go:embed embed/Dockerfile
var dockerfile []byte

//go:embed embed/entrypoint.sh
var entrypointSh []byte

// DefaultDockerfile returns the content of the embedded default Dockerfile.
func DefaultDockerfile() []byte {
	return dockerfile
}
