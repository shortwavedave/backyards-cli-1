package ale

import "fmt"

func (a *TCPAddr) String() string {
	return fmt.Sprintf("tcp://%s:%d", a.IP, a.Port)
}
