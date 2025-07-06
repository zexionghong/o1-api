package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// QueryBuilder 查询构建器
type QueryBuilder struct {
	db     *sql.DB
	dbx    *sqlx.DB
	table  string
	fields []string
	where  []string
	args   []interface{}
	orderBy string
	limit  int
	offset int
}

// NewQueryBuilder 创建查询构建器
func NewQueryBuilder(db *sql.DB, dbx *sqlx.DB, table string) *QueryBuilder {
	return &QueryBuilder{
		db:    db,
		dbx:   dbx,
		table: table,
		args:  make([]interface{}, 0),
	}
}

// Select 设置查询字段
func (qb *QueryBuilder) Select(fields ...string) *QueryBuilder {
	qb.fields = fields
	return qb
}

// Where 添加WHERE条件
func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	qb.where = append(qb.where, condition)
	qb.args = append(qb.args, args...)
	return qb
}

// WhereEqual 添加等值条件
func (qb *QueryBuilder) WhereEqual(field string, value interface{}) *QueryBuilder {
	return qb.Where(fmt.Sprintf("%s = ?", field), value)
}

// WhereIn 添加IN条件
func (qb *QueryBuilder) WhereIn(field string, values []interface{}) *QueryBuilder {
	if len(values) == 0 {
		return qb
	}
	
	placeholders := make([]string, len(values))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	
	condition := fmt.Sprintf("%s IN (%s)", field, strings.Join(placeholders, ","))
	return qb.Where(condition, values...)
}

// WhereBetween 添加BETWEEN条件
func (qb *QueryBuilder) WhereBetween(field string, start, end interface{}) *QueryBuilder {
	return qb.Where(fmt.Sprintf("%s BETWEEN ? AND ?", field), start, end)
}

// WhereNotNull 添加NOT NULL条件
func (qb *QueryBuilder) WhereNotNull(field string) *QueryBuilder {
	return qb.Where(fmt.Sprintf("%s IS NOT NULL", field))
}

// WhereNull 添加NULL条件
func (qb *QueryBuilder) WhereNull(field string) *QueryBuilder {
	return qb.Where(fmt.Sprintf("%s IS NULL", field))
}

// OrderBy 设置排序
func (qb *QueryBuilder) OrderBy(field string, direction ...string) *QueryBuilder {
	dir := "ASC"
	if len(direction) > 0 {
		dir = strings.ToUpper(direction[0])
	}
	qb.orderBy = fmt.Sprintf("%s %s", field, dir)
	return qb
}

// Limit 设置限制数量
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

// Offset 设置偏移量
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	return qb
}

// Paginate 设置分页
func (qb *QueryBuilder) Paginate(page, pageSize int) *QueryBuilder {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	
	qb.limit = pageSize
	qb.offset = (page - 1) * pageSize
	return qb
}

// BuildSelectQuery 构建SELECT查询
func (qb *QueryBuilder) BuildSelectQuery() (string, []interface{}) {
	// 构建字段列表
	fields := "*"
	if len(qb.fields) > 0 {
		fields = strings.Join(qb.fields, ", ")
	}
	
	// 构建基础查询
	query := fmt.Sprintf("SELECT %s FROM %s", fields, qb.table)
	
	// 添加WHERE条件
	if len(qb.where) > 0 {
		query += " WHERE " + strings.Join(qb.where, " AND ")
	}
	
	// 添加ORDER BY
	if qb.orderBy != "" {
		query += " ORDER BY " + qb.orderBy
	}
	
	// 添加LIMIT和OFFSET
	if qb.limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", qb.limit)
		if qb.offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", qb.offset)
		}
	}
	
	return query, qb.args
}

// BuildCountQuery 构建COUNT查询
func (qb *QueryBuilder) BuildCountQuery() (string, []interface{}) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", qb.table)
	
	// 添加WHERE条件
	if len(qb.where) > 0 {
		query += " WHERE " + strings.Join(qb.where, " AND ")
	}
	
	return query, qb.args
}

// QueryRow 执行单行查询
func (qb *QueryBuilder) QueryRow(ctx context.Context) *sql.Row {
	query, args := qb.BuildSelectQuery()
	return qb.db.QueryRowContext(ctx, query, args...)
}

// Query 执行多行查询
func (qb *QueryBuilder) Query(ctx context.Context) (*sql.Rows, error) {
	query, args := qb.BuildSelectQuery()
	return qb.db.QueryContext(ctx, query, args...)
}

// Count 执行计数查询
func (qb *QueryBuilder) Count(ctx context.Context) (int64, error) {
	query, args := qb.BuildCountQuery()
	var count int64
	err := qb.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

// Exists 检查记录是否存在
func (qb *QueryBuilder) Exists(ctx context.Context) (bool, error) {
	count, err := qb.Count(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// RepositoryHelper Repository助手
type RepositoryHelper struct {
	db  *sql.DB
	dbx *sqlx.DB
}

// NewRepositoryHelper 创建Repository助手
func NewRepositoryHelper(db *sql.DB, dbx *sqlx.DB) *RepositoryHelper {
	return &RepositoryHelper{
		db:  db,
		dbx: dbx,
	}
}

// NewQueryBuilder 创建查询构建器
func (h *RepositoryHelper) NewQueryBuilder(table string) *QueryBuilder {
	return NewQueryBuilder(h.db, h.dbx, table)
}

// GetByID 根据ID获取记录
func (h *RepositoryHelper) GetByID(ctx context.Context, table string, id int64, dest interface{}) error {
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", table)
	return h.dbx.GetContext(ctx, dest, query, id)
}

// GetByField 根据字段获取记录
func (h *RepositoryHelper) GetByField(ctx context.Context, table, field string, value interface{}, dest interface{}) error {
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = ?", table, field)
	return h.dbx.GetContext(ctx, dest, query, value)
}

// List 获取列表
func (h *RepositoryHelper) List(ctx context.Context, table string, offset, limit int, dest interface{}) error {
	query := fmt.Sprintf("SELECT * FROM %s ORDER BY created_at DESC LIMIT ? OFFSET ?", table)
	return h.dbx.SelectContext(ctx, dest, query, limit, offset)
}

// Count 获取总数
func (h *RepositoryHelper) Count(ctx context.Context, table string) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	var count int64
	err := h.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

// CountWithCondition 根据条件获取总数
func (h *RepositoryHelper) CountWithCondition(ctx context.Context, table, condition string, args ...interface{}) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", table, condition)
	var count int64
	err := h.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

// Insert 插入记录
func (h *RepositoryHelper) Insert(ctx context.Context, query string, args ...interface{}) (int64, error) {
	result, err := h.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// Update 更新记录
func (h *RepositoryHelper) Update(ctx context.Context, query string, args ...interface{}) (int64, error) {
	result, err := h.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Delete 删除记录
func (h *RepositoryHelper) Delete(ctx context.Context, table string, id int64) (int64, error) {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", table)
	result, err := h.db.ExecContext(ctx, query, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// SoftDelete 软删除记录
func (h *RepositoryHelper) SoftDelete(ctx context.Context, table string, id int64) (int64, error) {
	query := fmt.Sprintf("UPDATE %s SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL", table)
	result, err := h.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Exists 检查记录是否存在
func (h *RepositoryHelper) Exists(ctx context.Context, table, condition string, args ...interface{}) (bool, error) {
	count, err := h.CountWithCondition(ctx, table, condition, args...)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
