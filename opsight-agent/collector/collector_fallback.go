//go:build !linux && !windows

package collector

import "fmt"

// collect returns an error on unsupported platforms.
func collect() (*Metrics, error) {
	return nil, fmt.Errorf("unsupported platform: opsight-agent supports Linux and Windows only")
}
