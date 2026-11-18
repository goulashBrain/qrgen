package domain

// QRCode is the output of the use case: a named PNG image.
type QRCode struct {
	Name string // e.g. "001_simoncelli_leonardo.png"
	PNG  []byte
}
