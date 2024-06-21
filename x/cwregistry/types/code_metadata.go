package types

import (
	"errors"
	"net/url"
)

func (c CodeMetadata) Validate() error {
	if c.CodeId == 0 {
		return ErrNoSuchCode
	}
	if len(c.Schema) > 255 {
		return ErrSchemaTooLarge
	}
	for _, contact := range c.Contacts {
		if len(contact) > 255 {
			return errors.New("contact text is too long")
		}
	}
	if c.SourceBuilder != nil {
		if len(c.SourceBuilder.ContractName) > 255 {
			return errors.New("contract name is too long")
		}
		if len(c.SourceBuilder.Image) > 255 {
			return errors.New("image is too long")
		}
		if len(c.SourceBuilder.Tag) > 255 {
			return errors.New("tag is too long")
		}
	}
	if c.Source != nil {
		if len(c.Source.Repository) > 255 {
			return errors.New("repository url is too long")
		}
		if len(c.Source.Tag) > 255 {
			return errors.New("description is too long")
		}
		if len(c.Source.License) > 255 {
			return errors.New("license is too long")
		}
		if c.Source.Repository != "" {
			_, err := url.ParseRequestURI(c.Source.Repository)
			if err != nil {
				return errors.New("repository is not a valid URL")
			}
		}
	}
	return nil
}
