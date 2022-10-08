package chef

import (
	"context"

	"github.com/go-chef/chef"
)

func (s Service) GetDatabags(ctx context.Context) (interface{}, error) {
	environments, err := s.client.DataBags.List()
	if err != nil {
		return nil, err
	}

	return environments, nil
}

func (s Service) GetDatabagItems(ctx context.Context, name string) (*chef.DataBagListResult, error) {
	items, err := s.client.DataBags.ListItems(name)
	if err != nil {
		return items, err
	}

	return items, nil
}

func (s Service) GetDatabagItemContent(ctx context.Context, databag string, item string) (chef.DataBagItem, error) {
	contents, err := s.client.DataBags.GetItem(databag, item)
	if err != nil {
		return contents, err
	}

	return contents, nil
}
