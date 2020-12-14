package test

import (
	"github.com/stretchr/testify/suite"
)

type BaseTestSuite struct {
	suite.Suite
}

func (tb *BaseTestSuite) SetupSuite() {
}

func (tb *BaseTestSuite) TearDownSuite() {
}
