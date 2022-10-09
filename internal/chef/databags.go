package chef

import (
	"context"
	"errors"

	"github.com/go-chef/chef"
)

var (
	ErrDatabagNotFound     = errors.New("databag not found")
	ErrDatabagItemNotFound = errors.New("databag item not found")
)

func (s Service) GetDatabags(ctx context.Context) (interface{}, error) {
	databags, err := s.client.DataBags.List()
	if err != nil {
		return nil, err
	}

	return databags, nil
}

func (s Service) GetDatabagItems(ctx context.Context, name string) (*chef.DataBagListResult, error) {
	items, err := s.client.DataBags.ListItems(name)
	if err != nil {
		return items, ErrDatabagNotFound
	}

	return items, nil
}

func (s Service) GetDatabagItemContent(ctx context.Context, databag string, item string) (chef.DataBagItem, error) {
	contents, err := s.client.DataBags.GetItem(databag, item)
	if err != nil {
		return contents, ErrDatabagItemNotFound
	}

	return contents, nil
}
