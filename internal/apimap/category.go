package apimap

import (
	categoryv1 "trongcon-api/api/category/v1"
	"trongcon-api/internal/entity"
)

func CategoryToRes(c *entity.Category) categoryv1.CategoryRes {
	return categoryv1.CategoryRes{
		ID:        c.ID,
		Name:      c.Name,
		Icon:      c.Icon,
		Image:     c.Image,
		Status:    c.Status,
		Type:      c.Type,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}
