package apimap

import (
	"strings"

	articlev1 "trongcon-api/api/article/v1"
	"trongcon-api/internal/entity"
)

func articleAuthorName(u entity.User) string {
	if u.Name != "" {
		return u.Name
	}
	return strings.TrimSpace(u.FirstName + " " + u.LastName)
}

func ArticleToList(a *entity.Article) articlev1.ArticleListRes {
	out := articlev1.ArticleListRes{
		ID:         a.ID,
		Title:      a.Title,
		Subtitle:   a.Subtitle,
		Slug:       a.Slug,
		Thumbnail:  a.Thumbnail,
		Video:      a.Video,
		UserID:     a.UserID,
		CategoryID: a.CategoryID,
		Featured:   a.Featured,
		CreatedAt:  a.CreatedAt,
		UpdatedAt:  a.UpdatedAt,
	}
	if a.User.ID != 0 {
		out.UserEmail = a.User.Email
		out.AuthorName = articleAuthorName(a.User)
		out.AuthorProfilePicture = a.User.ProfilePicture
	}
	if a.Category.ID != 0 {
		out.CategoryName = a.Category.Name
	}
	return out
}

func ArticleToDetail(a *entity.Article) articlev1.ArticleDetailRes {
	out := articlev1.ArticleDetailRes{
		ID:         a.ID,
		Title:      a.Title,
		Subtitle:   a.Subtitle,
		Slug:       a.Slug,
		Thumbnail:  a.Thumbnail,
		Video:      a.Video,
		Content:    a.Content,
		UserID:     a.UserID,
		CategoryID: a.CategoryID,
		Featured:   a.Featured,
		CreatedAt:  a.CreatedAt,
		UpdatedAt:  a.UpdatedAt,
	}
	if a.User.ID != 0 {
		out.UserEmail = a.User.Email
		out.AuthorName = articleAuthorName(a.User)
		out.AuthorProfilePicture = a.User.ProfilePicture
	}
	if a.Category.ID != 0 {
		out.CategoryName = a.Category.Name
	}
	return out
}
