package sql

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/info"
	"github.com/urandom/readeef/db"
	"github.com/urandom/webfw"
)

type User struct {
	base.User
	logger webfw.Logger

	db *db.DB
}

func (u *User) Update() {
	if u.Err() != nil {
		return
	}

	i := u.Info()
	u.logger.Infof("Updating user %s\n", i.Login)

	tx, err := u.db.Begin()
	if err != nil {
		u.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(u.db.SQL("update_user"))
	if err != nil {
		u.Err(err)
		return
	}
	defer stmt.Close()

	res, err := stmt.Exec(i.FirstName, i.LastName, i.Email, i.Admin, i.Active, i.ProfileJSON, i.HashType, i.Salt, i.Hash, i.MD5API, i.Login)
	if err != nil {
		u.Err(err)
		return
	}

	if num, err := res.RowsAffected(); err == nil && num > 0 {
		if err := tx.Commit(); err != nil {
			u.Err(err)
		}

		return
	}

	stmt, err = tx.Preparex(u.db.SQL("create_user"))
	if err != nil {
		u.Err(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(i.Login, i.FirstName, i.LastName, i.Email, i.Admin, i.Active, i.ProfileJSON, i.HashType, i.Salt, i.Hash, i.MD5API)
	if err != nil {
		u.Err(err)
		return
	}

	if err := tx.Commit(); err != nil {
		u.Err(err)
	}

	return
}

func (u *User) Delete() {
	if u.Err() != nil {
		return
	}

	i := u.Info()
	u.logger.Infof("Deleting user %s\n", i.Login)

	if err := u.Validate(); err != nil {
		u.Err(err)
		return
	}

	tx, err := u.db.Begin()
	if err != nil {
		u.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(u.db.SQL("delete_user"))
	if err != nil {
		u.Err(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(i.Login)
	if err != nil {
		u.Err(err)
		return
	}

	if err := tx.Commit(); err != nil {
		u.Err(err)
	}
}

func (u *User) Feed(id info.FeedId) (uf content.UserFeed) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting user feed for user %s and feed %d\n", login, id)

	var i info.Feed
	if err := u.db.Get(&i, u.db.SQL("get_user_feed"), id, login); err != nil && err != sql.ErrNoRows {
		u.Err(err)
		return
	}

	uf.Info(i)

	return
}

func (u *User) AddFeed(f content.Feed) (uf content.UserFeed) {
	if u.Err() != nil {
		return
	}

	if err := f.Validate(); err != nil {
		u.Err(err)
		return
	}

	login := u.Info().Login
	i := f.Info()
	u.logger.Infof("Getting user feed for user %s and feed %d\n", login, i.Id)

	tx, err := u.db.Begin()
	if err != nil {
		u.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(u.db.SQL("create_user_feed"))
	if err != nil {
		u.Err(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(u.Info().Login, i.Id)
	if err != nil {
		u.Err(err)
		return
	}

	if err := tx.Commit(); err != nil {
		u.Err(err)
	}

	uf = u.Repo().UserFeed(u)
	uf.Info(i)

	return
}

func (u *User) AllFeeds() (uf []content.TaggedFeed) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting all feeds for user %s\n", login)

	var info []info.Feed
	if err := u.db.Select(&info, u.db.SQL("get_user_feeds"), login); err != nil {
		u.Err(err)
		return
	}

	uf = make([]content.TaggedFeed, len(info))
	for i := range info {
		uf[i] = u.Repo().TaggedFeed(u)
		uf[i].Info(info[i])
	}

	return
}

func (u *User) AllTaggedFeeds() (tf []content.TaggedFeed) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting all tagged feeds for user %s\n", login)

	var feedIdTags []feedIdTag

	if err := u.db.Select(&feedIdTags, u.db.SQL("get_user_feed_ids_tags"), login); err != nil {
		u.Err(err)
		return
	}

	tf = u.AllFeeds()
	if u.Err() != nil {
		return
	}

	feedMap := make(map[info.FeedId][]content.Tag)

	for _, tuple := range feedIdTags {
		tag := u.Repo().Tag(u)
		tag.Set(tuple.TagValue)
		feedMap[tuple.FeedId] = append(feedMap[tuple.FeedId], tag)
	}

	for i := range tf {
		tf[i].Tags(feedMap[tf[i].Info().Id])
	}

	return
}

func (u *User) Article(id info.ArticleId) (ua content.UserArticle) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting article '%d' for user %s\n", id, login)

	articles := getArticles(u, u.db, u.logger, u, "", "", "a.id = $2", "", []interface{}{id})

	if u.Err() == nil && len(articles) > 0 {
		return articles[0]
	}

	return
}

func (u *User) ArticlesById(ids ...info.ArticleId) (ua []content.UserArticle) {
	if u.Err() != nil || len(ids) == 0 {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting articles %q for user %s\n", ids, login)

	where := "("

	args := []interface{}{}
	index := 1
	for _, id := range ids {
		if index > 1 {
			where += ` OR `
		}

		where += fmt.Sprintf(`a.id = $%d`, index+1)
		args = append(args, id)
		index = len(args) + 1
	}

	where += ")"

	articles := getArticles(u, u.db, u.logger, u, "", "", where, "", args)
	ua = make([]content.UserArticle, len(articles))
	for i := range articles {
		ua[i] = articles[i]
	}

	return
}

func (u *User) AllUnreadArticleIds() (ids []info.ArticleId) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting unread article ids for user %s\n", login)

	if err := u.db.Select(&ids, u.db.SQL("get_all_unread_user_article_ids"), login); err != nil {
		u.Err(err)
		return
	}

	return
}

func (u *User) AllFavoriteIds() (ids []info.ArticleId) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting favorite article ids for user %s\n", login)

	if err := u.db.Select(&ids, u.db.SQL("get_all_favorite_user_article_ids"), login); err != nil {
		u.Err(err)
		return
	}

	return
}

func (u *User) ArticleCount() (c int64) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting article count for user %s\n", login)

	if err := u.db.Get(&c, u.db.SQL("get_user_article_count"), login); err != nil && err != sql.ErrNoRows {
		u.Err(err)
		return
	}

	return
}

func (u *User) Articles(paging ...int) (ua []content.UserArticle) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting articles for paging %q and user %s\n", paging, login)

	order := "read"

	articles := getArticles(u, u.db, u.logger, u, "", "", "", order, nil, paging...)
	ua = make([]content.UserArticle, len(articles))
	for i := range articles {
		ua[i] = articles[i]
	}

	return
}

func (u *User) UnreadArticles(paging ...int) (ua []content.UserArticle) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting unread articles for paging %q and user %s\n", paging, login)

	articles := getArticles(u, u.db, u.logger, u, "", "", "ar.article_id IS NULL", "", nil, paging...)
	ua = make([]content.UserArticle, len(articles))
	for i := range articles {
		ua[i] = articles[i]
	}

	return
}

func (u *User) ArticlesOrderedById(pivot info.ArticleId, paging ...int) (ua []content.UserArticle) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting articles order by id for paging %q and user %s\n", paging, login)

	u.SortingById()

	articles := getArticles(u, u.db, u.logger, u, "", "", "a.id > $2", "", []interface{}{pivot}, paging...)
	ua = make([]content.UserArticle, len(articles))
	for i := range articles {
		ua[i] = articles[i]
	}

	return
}

func (u *User) FavoriteArticles(paging ...int) (ua []content.UserArticle) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting favorite articles for paging %q and user %s\n", paging, login)

	articles := getArticles(u, u.db, u.logger, u, "", "", "af.article_id IS NOT NULL", "", nil, paging...)
	ua = make([]content.UserArticle, len(articles))
	for i := range articles {
		ua[i] = articles[i]
	}

	return
}

func (u *User) ReadBefore(date time.Time, read bool) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Marking user %s articles before %v as read: %v\n", login, date, read)

	tx, err := u.db.Begin()
	if err != nil {
		u.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(u.db.SQL("delete_all_user_articles_read_by_date"))
	if err != nil {
		u.Err(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(login, date)
	if err != nil {
		u.Err(err)
		return
	}

	if read {
		stmt, err = tx.Preparex(u.db.SQL("create_all_user_articles_read_by_date"))

		if err != nil {
			u.Err(err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(login, date)
		if err != nil {
			u.Err(err)
			return
		}
	}

	tx.Commit()
}

func (u *User) ReadAfter(date time.Time, read bool) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Marking user %s articles after %v as read: %v\n", login, date, read)

	tx, err := u.db.Begin()
	if err != nil {
		u.Err(err)
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(u.db.SQL("delete_newer_user_articles_read_by_date"))

	if err != nil {
		u.Err(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(login, date)
	if err != nil {
		u.Err(err)
		return
	}

	if read {
		stmt, err = tx.Preparex(u.db.SQL("create_newer_user_articles_read_by_date"))

		if err != nil {
			u.Err(err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(login, date)
		if err != nil {
			u.Err(err)
			return
		}
	}

	return
}

func (u *User) ScoredArticles(from, to time.Time, paging ...int) (sa []content.ScoredArticle) {
	if u.Err() != nil {
		return
	}

	login := u.Info().Login
	u.logger.Infof("Getting scored articles for paging %q and user %s\n", paging, login)

	order := "asco.score"
	if u.Order() == info.DescendingOrder {
		order = "asco.score DESC"
	}

	ua := getArticles(u, u.db, u.logger, u, "asco.score",
		"INNER JOIN articles_scores asco ON a.id = asco.article_id",
		"a.date > $2 AND a.date <= $3", order,
		[]interface{}{from, to}, paging...)

	sa = make([]content.ScoredArticle, len(ua))
	for i := range ua {
		sa[i] = u.Repo().ScoredArticle(u)
		sa[i].Info(ua[i].Info())
	}

	return sa
}

func (u *User) Tags() (tags []content.Tag) {
	if u.Err() != nil {
		return
	}

	return
}

func getArticles(u content.User, dbo *db.DB, logger webfw.Logger, sorting content.ArticleSorting, columns, join, where, order string, args []interface{}, paging ...int) (ua []content.UserArticle) {
	if u.Err() != nil {
		return
	}

	sql := dbo.SQL("get_article_columns")
	if columns != "" {
		sql += ", " + columns
	}

	sql += dbo.SQL("get_article_tables")
	if join != "" {
		sql += " " + join
	}

	sql += dbo.SQL("get_article_joins")

	args = append([]interface{}{u.Info().Login}, args...)
	if where != "" {
		sql += " AND " + where
	}

	sortingField := sorting.Field()
	sortingOrder := sorting.Order()

	fields := []string{}
	if order != "" {
		fields = append(fields, order)
	}
	switch sortingField {
	case info.SortById:
		fields = append(fields, "a.id")
	case info.SortByDate:
		fields = append(fields, "a.date")
	}
	if len(fields) > 0 {
		sql += " ORDER BY "

		sql += strings.Join(fields, ",")

		if sortingOrder == info.DescendingOrder {
			sql += " DESC"
		}
	}

	if len(paging) > 0 {
		limit, offset := pagingLimit(paging)

		sql += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
		args = append(args, limit, offset)
	}

	var info []info.Article
	if err := dbo.Select(&info, sql, args...); err != nil {
		u.Err(err)
		return
	}

	ua = make([]content.UserArticle, len(info))
	for i := range info {
		ua[i] = u.Repo().UserArticle(u)
		ua[i].Info(info[i])
	}

	return
}