package constants

const (
	Folder                 = ".gonx"
	BenchmarkFolderPath    = Folder + "/benchmarks"
	BundleAnalyserFile     = "bundle-benchmarks.json"
	BundleAnalyserFilePath = BenchmarkFolderPath + "/" + BundleAnalyserFile

	BuildAnalyserFile     = "build-benchmarks.json"
	BuildAnalyserFilePath = BenchmarkFolderPath + "/" + BuildAnalyserFile

	LintAnalyserFile     = "lint-benchmarks.json"
	LintAnalyserFilePath = BenchmarkFolderPath + "/" + LintAnalyserFile

	TestAnalyserFile     = "test-benchmarks.json"
	TestAnalyserFilePath = BenchmarkFolderPath + "/" + TestAnalyserFile
)
