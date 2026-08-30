package progressbar

type ProgressBar interface {
	Stop()
	Run()
	Set(int)
}
