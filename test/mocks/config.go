package mocks

//Mock for Writer
type MockWriter struct {
	write func(p []byte) (n int, err error)
}

func NewMockWriter() *MockWriter {
	return &MockWriter{}
}

func (m *MockWriter) Write(p []byte) (n int, err error) {
	return m.write(p)
}
func (m *MockWriter) MockWrite(mocked func(p []byte) (n int, err error)) {
	m.write = mocked
}

//Mock for ConfigReader
type MockConfigReader struct {
	getString func(string) string
}

func NewMockConfigReader() *MockConfigReader {
	return &MockConfigReader{}
}

func (m *MockConfigReader) GetString(key string) string {
	return m.getString(key)
}

func (m *MockConfigReader) MockGetString(mocked func(string) string) {
	m.getString = mocked
}
