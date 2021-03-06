package content

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/urandom/readeef/content/data"
)

type ArticleSorting interface {
	// Resets the sorting
	DefaultSorting() ArticleSorting

	// Sorts by content id, if available
	SortingById() ArticleSorting
	// Sorts by date, if available
	SortingByDate() ArticleSorting
	// Reverse the order
	Reverse() ArticleSorting

	// Returns the current field
	Field(f ...data.SortingField) data.SortingField

	// Returns the order, as set by Reverse()
	Order(o ...data.Order) data.Order
}

type ArticleRepo interface {
	Articles(paging ...int) []UserArticle
	UnreadArticles(paging ...int) []UserArticle

	ReadBefore(date time.Time, read bool)

	ScoredArticles(from, to time.Time, paging ...int) []ScoredArticle
}

type ArticleSearch interface {
	Highlight(highlight ...string) string
	Query(query string, index bleve.Index, paging ...int) []UserArticle
}

type Article interface {
	Error
	RepoRelated

	fmt.Stringer
	json.Marshaler

	Data(data ...data.Article) data.Article

	Validate() error

	Update()

	Format(templateDir, readabilityKey string) data.ArticleFormatting
}

type ScoredArticle interface {
	Article

	Scores() ArticleScores
}

type UserArticle interface {
	Article
	UserRelated

	Read(read bool)
	Favorite(favorite bool)
}

type ArticleScores interface {
	Error
	RepoRelated

	fmt.Stringer

	Data(data ...data.ArticleScores) data.ArticleScores

	Validate() error
	Calculate() int64

	Update()
}
