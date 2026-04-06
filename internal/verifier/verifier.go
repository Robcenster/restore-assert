package verifier

import (
	"context"
	"fmt"

	"github.com/Robcenster/restore-assert/internal/config"
	"github.com/Robcenster/restore-assert/internal/repository"
)

type Verifier struct {
	source repository.DBRepository
}

func NewVerifier(source repository.DBRepository) *Verifier {
	return &Verifier{source: source}
}

func (v *Verifier) RunAssert(ctx context.Context, assert config.AssertConfig) (bool, error) {
	actualRaw, err := v.source.ExecuteQuery(ctx, assert.Query)
	if err != nil {
		if assert.Type == "error" {
			return true, err
		}
		return false, err
	} else if assert.Type == "noerror" {
		return true, nil
	}

	expectedStr := fmt.Sprintf("%v", assert.Expected)

	success, err := Compare(assert.Type, actualRaw, expectedStr, assert.Operator)
	if err != nil {
		return false, err
	}

	return success, nil
}
