package repositories

import (
	"context"
	"fmt"
	"go_project_structure/common_pkg/pagination/cursor_pagination"
	"go_project_structure/common_pkg/pagination/offset_pagination"
	"go_project_structure/common_pkg/pagination/seek_pagination"
	"go_project_structure/internal/infrastructure/models"
	"time"
)

// offset based pagination
func (u *UserRepositoryImpl) ListUsersOffsetPagination(ctx context.Context, p offset_pagination.Params) ([]*models.User, int64, error) {
	log := u.userRepositoryLog.WithContext(ctx).Method("GetAllByOffsetPagination")

	// step 1: prepare the query
	query := "SELECT id, name, email, created_at, updated_at FROM users"
	countQuery := "SELECT COUNT(*) FROM users WHERE deleted_at IS NULL"
	args := []interface{}{}

	query += " WHERE deleted_at IS NULL"
	query += " ORDER BY created_at DESC"

	if p.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, p.Limit)
	}

	if p.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, p.Offset)
	}

	// step 2: execute the query

	// fetching users
	var users []*models.User
	usersErr := u.db.WithContext(ctx).Raw(query, args...).Scan(&users).Error
	if usersErr != nil {
		log.Errorf("Error fetching all users: %v\n", usersErr)
		return nil, 0, usersErr
	}

	// count users
	var totalUsers int64
	cntErr := u.db.WithContext(ctx).Raw(countQuery).Scan(&totalUsers).Error
	if cntErr != nil {
		log.Errorf("Error counting all users: %v\n", cntErr)
		return nil, 0, cntErr
	}

	// step 3: return the result
	log.Infof("Fetched users successfully from repository.")
	return users, totalUsers, nil
}

// cursor based pagination
func (u *UserRepositoryImpl) ListUsersCursorPagination(ctx context.Context, p cursor_pagination.Params) ([]*models.User, error) {
	log := u.userRepositoryLog.WithContext(ctx).Method("ListUsersCursorPagination")

	// step 1: prepare the query
	query := "SELECT id, name, email, created_at, updated_at FROM users"
	// countQuery := "SELECT COUNT(*) FROM users WHERE deleted_at IS NULL"
	args := []interface{}{}

	query += " WHERE deleted_at IS NULL"
	// query += " ORDER BY created_at DESC"

	if p.Cursor != nil {
		switch p.Direction {
		case cursor_pagination.DirectionNext:
			query += " AND ((created_at < ?) OR (created_at = ? AND id < ?))"
			args = append(args, p.Cursor.CreatedAt, p.Cursor.CreatedAt, p.Cursor.ID)
		case cursor_pagination.DirectionPrev:
			query += " AND ((created_at > ?) OR (created_at = ? AND id > ?))"
			args = append(args, p.Cursor.CreatedAt, p.Cursor.CreatedAt, p.Cursor.ID)
		}
	}

	if p.Direction == cursor_pagination.DirectionNext {
		query += " ORDER BY created_at DESC, id DESC"
	} else {
		query += " ORDER BY created_at ASC, id ASC"
	}

	// step 2: execute the query
	var users []*models.User
	usersErr := u.db.WithContext(ctx).Raw(query, args...).Scan(&users).Error
	if usersErr != nil {
		log.Errorf("Error fetching all users: %v\n", usersErr)
		return nil, usersErr
	}
	// step 3: return the result
	log.Infof("Fetched users successfully from repository.")
	return users, nil
}

// seek based pagination
func (u *UserRepositoryImpl) ListUsersSeekPagination(ctx context.Context, params seek_pagination.Params) (seek_pagination.RailResult[models.User], error) {

	if params.Cursor == nil {
		return u.seekFirstPage(ctx, params.Limit)
	}

	if params.Direction == seek_pagination.DirectionPrev {
		return u.seekPrevPage(ctx, params.Cursor, params.Limit)
	}

	return u.seekNextPage(ctx, params.Cursor, params.Limit)
}

func (u *UserRepositoryImpl) CountUsersNewSince(ctx context.Context, since time.Time, sinceID uint) (int64, error) {
	const query = `
		SELECT COUNT(*)
		FROM   users
		WHERE  deleted_at IS NULL
		  AND  (created_at > ? OR (created_at = ? AND id > ?))`

	var count int64
	if err := u.db.WithContext(ctx).Raw(query, since, since, sinceID).Scan(&count).Error; err != nil {
		return 0, fmt.Errorf("count users new since: %w", err)
	}
	return count, nil
}

// seek based pagination helpers
func (u *UserRepositoryImpl) seekFirstPage(ctx context.Context, limit int) (seek_pagination.RailResult[models.User], error) {
	log := u.userRepositoryLog.WithContext(ctx).Method("seekFirstPage")

	// step 1: prepare the query
	query := fmt.Sprintf(`
		SELECT id, created_at, updated_at, deleted_at, name, email, password
		FROM   users
		WHERE  deleted_at IS NULL
		ORDER  BY created_at DESC, id DESC
		LIMIT  ?`)

	args := []interface{}{}
	args = append(args, limit+1)

	// step 2: execute the query
	var users []models.User
	usersErr := u.db.WithContext(ctx).Raw(query, args...).Scan(&users).Error
	if usersErr != nil {
		log.Errorf("Error fetching all users: %v\n", usersErr)
		return seek_pagination.RailResult[models.User]{}, fmt.Errorf("rawQueryUsers: %w", usersErr)

	}

	hasMore := len(users) > limit
	if hasMore {
		users = users[:limit]
	}

	return seek_pagination.RailResult[models.User]{
		OldItems:   users,
		HasMoreOld: hasMore,
	}, nil
}

func (u *UserRepositoryImpl) seekNextPage(ctx context.Context, c *seek_pagination.Cursor, limit int) (seek_pagination.RailResult[models.User], error) {
	log := u.userRepositoryLog.Method("seekNextPage").WithContext(ctx)

	result := seek_pagination.RailResult[models.User]{}
	remaining := limit

	// new rail
	newItems, hasMoreNew, err := u.seekFetchNewRail(ctx, c, remaining)
	if err != nil {
		log.Errorf("Error fetching all users: %v\n", err)
		return result, err
	}
	result.NewItems = newItems
	result.HasMoreNew = hasMoreNew
	remaining -= len(newItems)

	// history rail — fill leftover slots
	if remaining > 0 {
		oldItems, hasMoreOld, err := u.seekFetchHistoryRail(ctx, c, remaining)
		if err != nil {
			log.Errorf("Error fetching all users: %v\n", err)
			return result, err
		}
		result.OldItems = oldItems
		result.HasMoreOld = hasMoreOld
	}

	return result, nil

}

func (u *UserRepositoryImpl) seekPrevPage(ctx context.Context, c *seek_pagination.Cursor, limit int) (seek_pagination.RailResult[models.User], error) {
	log := u.userRepositoryLog.Method("seekPrevPage").WithContext(ctx)

	query := fmt.Sprintf(`
			SELECT id, created_at, updated_at, deleted_at, name, email, password
			FROM   users
			WHERE  deleted_at IS NULL
			AND  (created_at > ? OR (created_at = ? AND id > ?))
			AND  (created_at < ? OR (created_at = ? AND id <= ?))
			ORDER  BY created_at ASC, id ASC
			LIMIT  ?`)

	args := []any{
		c.LastCreatedAt, c.LastCreatedAt, c.LastID,
		c.AnchorCreatedAt, c.AnchorCreatedAt, c.AnchorID,
	}
	args = append(args, limit+1)

	var users []models.User
	usersErr := u.db.WithContext(ctx).Raw(query, args...).Scan(&users).Error
	if usersErr != nil {
		log.Errorf("Error fetching all users: %v\n", usersErr)
		return seek_pagination.RailResult[models.User]{}, fmt.Errorf("rawQueryUsers: %w", usersErr)

	}

	hasMore := len(users) > limit
	if hasMore {
		users = users[:limit]
	}
	reverseUsers(users)

	return seek_pagination.RailResult[models.User]{
		OldItems:   users,
		HasMoreOld: hasMore,
	}, nil
}

func (u *UserRepositoryImpl) seekFetchNewRail(ctx context.Context, c *seek_pagination.Cursor, limit int) ([]models.User, bool, error) {
	log := u.userRepositoryLog.WithContext(ctx).Method("seekFetchNewRail")

	query := fmt.Sprintf(`
			SELECT id, created_at, updated_at, deleted_at, name, email, password
			FROM   users
			WHERE  deleted_at IS NULL
			AND  (created_at > ? OR (created_at = ? AND id > ?))
			ORDER  BY created_at ASC, id ASC
			LIMIT  ?
			OFFSET ?`)

	args := []any{c.AnchorCreatedAt, c.AnchorCreatedAt, c.AnchorID}
	args = append(args, limit+1, c.NewRailOffset)

	var users []models.User
	usersErr := u.db.WithContext(ctx).Raw(query, args...).Scan(&users).Error
	if usersErr != nil {
		log.Errorf("Error fetching all users: %v\n", usersErr)
		return nil, false, fmt.Errorf("rawQueryUsers: %w", usersErr)

	}

	hasMore := len(users) > limit
	if hasMore {
		users = users[:limit]
	}
	return users, hasMore, nil

}

func (u *UserRepositoryImpl) seekFetchHistoryRail(ctx context.Context, c *seek_pagination.Cursor, limit int) ([]models.User, bool, error) {
	log := u.userRepositoryLog.WithContext(ctx).Method("seekFetchHistoryRail")

	query := fmt.Sprintf(`
		SELECT id, created_at, updated_at, deleted_at, name, email, password
		FROM   users
		WHERE  deleted_at IS NULL
		  AND  (created_at < ? OR (created_at = ? AND id < ?))
		  AND  (created_at < ? OR (created_at = ? AND id <= ?))
		ORDER  BY created_at DESC, id DESC
		LIMIT  ?`)

	args := []any{
		c.LastCreatedAt, c.LastCreatedAt, c.LastID,
		c.AnchorCreatedAt, c.AnchorCreatedAt, c.AnchorID,
	}
	args = append(args, limit+1)

	var users []models.User
	usersErr := u.db.WithContext(ctx).Raw(query, args...).Scan(&users).Error
	if usersErr != nil {
		log.Errorf("Error fetching all users: %v\n", usersErr)
		return nil, false, fmt.Errorf("rawQueryUsers: %w", usersErr)

	}

	hasMore := len(users) > limit
	if hasMore {
		users = users[:limit]
	}
	return users, hasMore, nil

}

func reverseUsers(s []models.User) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
