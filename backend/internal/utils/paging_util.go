package utils

import (
	"reflect"
	"strconv"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PaginationResponse struct {
	TotalPages   int64 `json:"totalPages"`
	TotalItems   int64 `json:"totalItems"`
	CurrentPage  int   `json:"currentPage"`
	ItemsPerPage int   `json:"itemsPerPage"`
}

type SortedPaginationRequest struct {
	Pagination struct {
		Page  int `form:"pagination[page]"`
		Limit int `form:"pagination[limit]"`
	} `form:"pagination"`
	Sort struct {
		Column    string `form:"sort[column]"`
		Direction string `form:"sort[direction]"`
	} `form:"sort"`
}

func PaginateAndSort(sortedPaginationRequest SortedPaginationRequest, query *gorm.DB, result interface{}) (PaginationResponse, error) {
	pagination := sortedPaginationRequest.Pagination
	sort := sortedPaginationRequest.Sort

	capitalizedSortColumn := CapitalizeFirstLetter(sort.Column)

	sortField, sortFieldFound := reflect.TypeOf(result).Elem().Elem().FieldByName(capitalizedSortColumn)
	isSortable, _ := strconv.ParseBool(sortField.Tag.Get("sortable"))

	if sort.Direction == "" || (sort.Direction != "asc" && sort.Direction != "desc") {
		sort.Direction = "asc"
	}

	if sortFieldFound && isSortable {
		columnName := CamelCaseToSnakeCase(sort.Column)
		query = query.Clauses(clause.OrderBy{
			Columns: []clause.OrderByColumn{
				{Column: clause.Column{Name: columnName}, Desc: sort.Direction == "desc"},
			},
		})
	}

	return Paginate(pagination.Page, pagination.Limit, query, result)
}

func Paginate(page int, pageSize int, query *gorm.DB, result interface{}) (PaginationResponse, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 20
	} else if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	var totalItems int64
	if err := query.Count(&totalItems).Error; err != nil {
		return PaginationResponse{}, err
	}

	if err := query.Offset(offset).Limit(pageSize).Find(result).Error; err != nil {
		return PaginationResponse{}, err
	}

	totalPages := (totalItems + int64(pageSize) - 1) / int64(pageSize)
	if totalItems == 0 {
		totalPages = 1
	}

	return PaginationResponse{
		TotalPages:   totalPages,
		TotalItems:   totalItems,
		CurrentPage:  page,
		ItemsPerPage: pageSize,
	}, nil
}
