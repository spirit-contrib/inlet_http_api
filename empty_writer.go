package main

type EmptyWriter struct {
}

func (p *EmptyWriter) Write(v []byte) (n int, err error) {
	return
}
