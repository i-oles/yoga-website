package generator

type Token interface {
	Generate(length int) (string, error)
}
