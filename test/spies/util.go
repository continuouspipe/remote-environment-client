package spies

//Mock for QuestionPrompt
type SpyQuestionPrompt struct {
	Spy
	readString       func(q string) string
	applyDefault     func(question string, predef string) string
	repeatIfEmpty    func(question string) string
	repeatUntilValid func(question string, isValid func(string) (bool, error)) string
}

func NewSpyQuestionPrompt() *SpyQuestionPrompt {
	return &SpyQuestionPrompt{}
}

func (qp *SpyQuestionPrompt) ReadString(q string) string {
	args := make(Arguments)
	args["q"] = q

	function := &Function{Name: "ReadString", Arguments: args}
	qp.calledFunctions = append(qp.calledFunctions, *function)
	return qp.readString(q)
}

func (qp *SpyQuestionPrompt) ApplyDefault(question string, predef string) string {
	args := make(Arguments)
	args["question"] = question
	args["predef"] = predef

	function := &Function{Name: "ApplyDefault", Arguments: args}
	qp.calledFunctions = append(qp.calledFunctions, *function)
	return qp.applyDefault(question, predef)
}

func (qp *SpyQuestionPrompt) RepeatIfEmpty(question string) string {
	args := make(Arguments)
	args["question"] = question

	function := &Function{Name: "RepeatIfEmpty", Arguments: args}
	qp.calledFunctions = append(qp.calledFunctions, *function)
	return qp.repeatIfEmpty(question)
}

func (qp *SpyQuestionPrompt) RepeatUntilValid(question string, isValid func(string) (bool, error)) string {
	args := make(Arguments)
	args["question"] = question
	args["isValid"] = isValid

	function := &Function{Name: "RepeatUntilValid", Arguments: args}
	qp.calledFunctions = append(qp.calledFunctions, *function)
	return qp.repeatUntilValid(question, isValid)
}

func (qp *SpyQuestionPrompt) MockReadString(mocked func(q string) string) {
	qp.readString = mocked
}

func (qp *SpyQuestionPrompt) MockApplyDefault(mocked func(question string, predef string) string) {
	qp.applyDefault = mocked
}

func (qp *SpyQuestionPrompt) MockRepeatIfEmpty(mocked func(question string) string) {
	qp.repeatIfEmpty = mocked
}

func (qp *SpyQuestionPrompt) MockRepeatUntilValid(mocked func(question string, isValid func(string) (bool, error)) string) {
	qp.repeatUntilValid = mocked
}
