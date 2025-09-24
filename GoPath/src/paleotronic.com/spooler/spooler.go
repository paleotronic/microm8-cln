package spooler

type PrintSpooler interface {
	SpoolPDF(filename string) error
}
