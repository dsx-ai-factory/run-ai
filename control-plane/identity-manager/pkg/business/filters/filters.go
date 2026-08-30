package business

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/generated/pkg/api"
	iamModels "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"
)

type (
	Field    string
	Operator string
)

const (
	FieldUsername     Field = "username"
	FieldCreatedBy    Field = "createdBy"
	FieldCreationTime Field = "creationTime"
	FieldLastLogin    Field = "lastLogin"
	FieldLastUpdated  Field = "lastUpdated"
	FieldIsLocal      Field = "isLocal"
	FieldFirstName    Field = "firstName"
	FieldLastName     Field = "lastName"
	FieldFullName     Field = "fullName"

	OperatorEqual              Operator = "="
	OperatorExact              Operator = "=="
	OperatorGreaterThan        Operator = ">"
	OperatorGreaterThanOrEqual Operator = ">="
	OperatorLessThan           Operator = "<"
	OperatorLessThanOrEqual    Operator = "<="
	OperatorContains           Operator = "=@"

	RegexPattern string = "^(username|createdBy|lastLogin|creationTime|lastUpdated|isLocal|firstName|lastName|fullName)(==|>=|<=|=@)(.+)$"

	UsersMinimalRegexPattern    = `^username(==|=@)(.+)$`
	UsersMinimalMaxIDs          = 25
	UsersMinimalMaxFilterBy     = 25
	UsersMinimalMinSearchLength = 2
)

var OperatorMap = map[string]Operator{
	string(OperatorGreaterThanOrEqual): OperatorGreaterThan,
	string(OperatorLessThanOrEqual):    OperatorLessThan,
	string(OperatorExact):              OperatorEqual,
}

func ParseFilters(filterBy []string, sortBy api.UsersFilterSortFields, sortOrder string, offset int32, limit int32, search string) (*iamModels.UserFilters, error) {
	userFilters := &iamModels.UserFilters{}

	for _, filter := range filterBy {
		field, operator, value, err := ValidateFilter(filter)
		if err != nil {
			return nil, err
		}

		operator = ConvertOperator(operator)

		if isDateField(field) {
			value, err = utils.Iso8601ToEpoch(value)
			if err != nil {
				return nil, err
			}
		}
		if err := AssignField(userFilters, field, operator, value); err != nil {
			return nil, err
		}
	}

	filters, err := enrichFiltersWithPaginationProps(userFilters, sortBy, limit, offset, sortOrder, search)
	if err != nil {
		return filters, err
	}
	return userFilters, nil
}

func enrichFiltersWithPaginationProps(userFilters *iamModels.UserFilters, sortBy api.UsersFilterSortFields, limit int32, offset int32, sortOrder string, search string) (*iamModels.UserFilters, error) {
	if sortBy != "" {
		if sortBy == "user" {
			sortBy = api.USERSFILTERSORTFIELDS_USERNAME
		}

		if !sortBy.IsValid() {
			return nil, fmt.Errorf("invalid sort by: %s", sortBy)
		}
		sortByStr := string(sortBy)
		userFilters.SortBy = &sortByStr
	}
	if limit > 0 {
		userFilters.Max = int32ToIntPtr(limit)
	}
	if offset > 0 {
		userFilters.Offset = int32ToIntPtr(offset)
	}
	if sortOrder != "" && sortOrder != "asc" && sortOrder != "desc" {
		return nil, fmt.Errorf("invalid sort order: %s", sortOrder)
	}
	userFilters.SortOrder = &sortOrder
	if search != "" {
		userFilters.Search = &search
	}
	return nil, nil
}

func ValidateFilter(filter string) (Field, Operator, string, error) {
	pattern := regexp.MustCompile(RegexPattern)
	matches := pattern.FindStringSubmatch(filter)

	if len(matches) < 4 {
		return "", "", "", fmt.Errorf("invalid filter: %s", filter)
	}

	field := Field(matches[1])
	operator := Operator(matches[2])
	value := matches[3]

	if isDateField(field) {
		return validateDatesFilter(field, operator, value)
	} else if isTextField(field) {
		return validateTextFilters(field, operator, value)
	} else if isBooleanField(field) {
		return validateBooleanFilters(field, operator, value)
	} else {
		return "", "", "", fmt.Errorf("invalid filter: %s", filter)
	}
}

func ConvertOperator(operator Operator) Operator {
	if op, exists := OperatorMap[string(operator)]; exists {
		return op
	}
	return operator
}

func AssignField(userFilters *iamModels.UserFilters, field Field, operator Operator, value string) error {
	assignments := map[Field]func(){
		FieldUsername:  func() { userFilters.Username = &value },
		FieldCreatedBy: func() { userFilters.CreatedBy = &value },
		FieldFirstName: func() { userFilters.FullName = &value },
		FieldLastName:  func() { userFilters.FullName = &value },
		FieldFullName:  func() { userFilters.FullName = &value },
		FieldCreationTime: func() {
			creationFilter := string(operator) + value
			userFilters.CreationFilter = &creationFilter
		},
		FieldLastLogin: func() {
			loginFilter := string(operator) + value
			userFilters.LoginFilter = &loginFilter
		},
		FieldLastUpdated: func() {
			updateFilter := string(operator) + value
			userFilters.UpdateFilter = &updateFilter
		},
		FieldIsLocal: func() {
			runaisLocal := "runai.is_local:false"
			if value == "true" {
				runaisLocal = "runai.is_local:true"
			}
			if userFilters.Q == nil {
				userFilters.Q = &runaisLocal
			} else {
				*userFilters.Q = *userFilters.Q + " " + runaisLocal
			}
		},
	}
	if assignFunc, exists := assignments[field]; exists {
		assignFunc()
		return nil
	}
	return fmt.Errorf("unsupported field: %s", field)
}

func isDateField(field Field) bool {
	return field == FieldCreationTime || field == FieldLastLogin || field == FieldLastUpdated
}

func isBooleanField(field Field) bool {
	return field == FieldIsLocal
}

func isTextField(field Field) bool {
	return field == FieldUsername || field == FieldCreatedBy || field == FieldFirstName || field == FieldLastName || field == FieldFullName
}

func validateTextFilters(field Field, operator Operator, value string) (Field, Operator, string, error) {
	if operator != OperatorContains {
		return "", "", "", fmt.Errorf("invalid operator %s for boolean field %s", operator, field)
	}
	return field, operator, value, nil
}

func validateBooleanFilters(field Field, operator Operator, value string) (Field, Operator, string, error) {
	if operator != OperatorExact {
		return "", "", "", fmt.Errorf("invalid operator %s for boolean field %s", operator, field)
	} else if value != "true" && value != "false" {
		return "", "", "", fmt.Errorf("invalid value %s for boolean field %s", value, field)
	}
	return field, operator, value, nil
}

func validateDatesFilter(field Field, operator Operator, value string) (Field, Operator, string, error) {
	if operator != OperatorGreaterThanOrEqual && operator != OperatorLessThanOrEqual && operator != OperatorExact {
		return "", "", "", fmt.Errorf("invalid operator %s for date field %s, only '>=' or '<=' allowed", operator, field)
	}
	return field, operator, value, nil
}

func int32ToIntPtr(val int32) *int {
	intVal := int(val)
	return &intVal
}

func ParseUsersMinimalFilters(filterBy []string, limit int32, search string, ids []string) (*iamModels.UserFilters, []string, error) {
	userFilters := &iamModels.UserFilters{}

	if search == "" && len(ids) == 0 && len(filterBy) == 0 {
		return nil, nil, fmt.Errorf("at least one of search, id, or filterBy is required")
	}

	if search != "" && len(search) < UsersMinimalMinSearchLength {
		return nil, nil, fmt.Errorf("search must be at least %d characters", UsersMinimalMinSearchLength)
	}

	if len(ids) > UsersMinimalMaxIDs {
		return nil, nil, fmt.Errorf("too many ids: at most %d ids are allowed, got %d", UsersMinimalMaxIDs, len(ids))
	}

	for i, id := range ids {
		if strings.TrimSpace(id) == "" {
			return nil, nil, fmt.Errorf("invalid id at position %d: must not be blank", i)
		}
	}

	if len(filterBy) > UsersMinimalMaxFilterBy {
		return nil, nil, fmt.Errorf("too many filterBy entries: at most %d are allowed, got %d", UsersMinimalMaxFilterBy, len(filterBy))
	}

	pattern := regexp.MustCompile(UsersMinimalRegexPattern)
	var bulkUsernamesExact []string
	for _, filter := range filterBy {
		matches := pattern.FindStringSubmatch(filter)
		if len(matches) < 3 {
			return nil, nil, fmt.Errorf("invalid filter: %s. Only username field with == or =@ operators is supported", filter)
		}
		operator := matches[1]
		value := matches[2]
		if len(filterBy) > 1 {
			if operator != string(OperatorExact) {
				return nil, nil, fmt.Errorf("multiple filterBy entries require the == operator; got %q", operator)
			}
			bulkUsernamesExact = append(bulkUsernamesExact, value)
		} else {
			userFilters.Username = &value
			if operator == string(OperatorExact) {
				exact := true
				userFilters.UsernameExact = &exact
			}
		}
	}

	if limit > 0 {
		userFilters.Max = int32ToIntPtr(limit)
	}
	if search != "" {
		userFilters.Search = &search
		minimalSearch := true
		userFilters.MinimalSearch = &minimalSearch
	}
	return userFilters, bulkUsernamesExact, nil
}
