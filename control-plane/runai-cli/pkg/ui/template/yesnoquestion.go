package template

import (
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/ui"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/ui/choises"
)

const (
	answerYes = "Yes"
	answerNo  = "No"
)

func YesNoQuestion(question string) (bool, error) {
	result, err := choises.ExecuteTea(question, "", []string{answerYes, answerNo})
	if err != nil && err.Error() != ui.ErrorNoChoiceWasMade {
		return false, err
	}

	if result != answerYes {
		return false, nil
	}
	return true, nil
}
