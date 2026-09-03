/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package template

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/choises"
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
