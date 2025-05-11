package repo

import (
	"database/sql"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDBPath = "test.db"

func setupTestDB(t *testing.T) *Repo {
	t.Helper()

	// Удаляем старую тестовую БД, если есть
	_ = os.Remove(testDBPath)

	db, err := sql.Open("sqlite3", testDBPath)
	require.NoError(t, err)

	err = createTables(db)
	require.NoError(t, err)

	return &Repo{db: db}
}

func cleanupTestDB(t *testing.T) {
	t.Helper()
	err := os.Remove(testDBPath)
	require.NoError(t, err)
}

func TestUserOperations(t *testing.T) {
	repo := setupTestDB(t)
	defer cleanupTestDB(t)

	// Тест InsertUser
	t.Run("InsertUser", func(t *testing.T) {
		user := User{
			Username: "testuser",
			Password: "testpass",
		}

		err := repo.InsertUser(user)
		assert.NoError(t, err)

		// Попытка вставить того же пользователя снова
		err = repo.InsertUser(user)
		assert.Error(t, err)
	})

	// Тест GetUser
	t.Run("GetUser", func(t *testing.T) {
		user := User{
			Username: "getuser",
			Password: "getpass",
		}

		err := repo.InsertUser(user)
		require.NoError(t, err)

		// Корректные данные
		foundUser, err := repo.GetUser(user.Username, user.Password)
		assert.NoError(t, err)
		assert.Equal(t, user.Username, foundUser.Username)

		// Неправильный пароль
		_, err = repo.GetUser(user.Username, "wrongpass")
		assert.Error(t, err)

		// Несуществующий пользователь
		_, err = repo.GetUser("nonexistent", "pass")
		assert.Error(t, err)
	})

	// Тест Authenticate
	t.Run("Authenticate", func(t *testing.T) {
		user := User{
			Username: "authuser",
			Password: "authpass",
		}

		err := repo.InsertUser(user)
		require.NoError(t, err)

		// Корректные данные
		authenticated, err := repo.Authenticate(user.Username, user.Password)
		assert.NoError(t, err)
		assert.True(t, authenticated)

		// Неправильный пароль
		authenticated, err = repo.Authenticate(user.Username, "wrongpass")
		assert.NoError(t, err)
		assert.False(t, authenticated)

		// Несуществующий пользователь
		authenticated, err = repo.Authenticate("nonexistent", "pass")
		assert.Error(t, err)
		assert.False(t, authenticated)
	})
}

func TestExpressionOperations(t *testing.T) {
	repo := setupTestDB(t)
	defer cleanupTestDB(t)

	// Создаем тестового пользователя
	user := User{
		Username: "expruser",
		Password: "exprpass",
	}
	err := repo.InsertUser(user)
	require.NoError(t, err)

	// Тест CreateExpression
	t.Run("CreateExpression", func(t *testing.T) {
		expr := &Expression{
			Username:   user.Username,
			Expression: "2 + 2",
			Status:     "pending",
		}

		err := repo.CreateExpression(expr)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, expr.ID)
	})

	// Тест GetExpressionByID и UpdateExpressionResult
	t.Run("GetAndUpdateExpression", func(t *testing.T) {
		expr := &Expression{
			Username:   user.Username,
			Expression: "3 * 3",
			Status:     "pending",
		}

		err := repo.CreateExpression(expr)
		require.NoError(t, err)

		// Получаем выражение
		foundExpr, err := repo.GetExpressionByID(expr.ID)
		assert.NoError(t, err)
		assert.Equal(t, expr.ID, foundExpr.ID)
		assert.Equal(t, expr.Expression, foundExpr.Expression)
		assert.Equal(t, 0, foundExpr.Result)
		assert.Equal(t, "pending", foundExpr.Status)

		// Обновляем результат
		err = repo.UpdateExpressionResult(expr.ID, 9, "completed")
		assert.NoError(t, err)

		// Проверяем обновление
		updatedExpr, err := repo.GetExpressionByID(expr.ID)
		assert.NoError(t, err)
		assert.Equal(t, 9, updatedExpr.Result)
		assert.Equal(t, "completed", updatedExpr.Status)
	})

	// Тест GetExpressions
	t.Run("GetExpressions", func(t *testing.T) {
		// Создаем несколько выражений
		expr1 := &Expression{
			Username:   user.Username,
			Expression: "5 + 5",
			Status:     "completed",
		}
		expr2 := &Expression{
			Username:   user.Username,
			Expression: "10 - 5",
			Status:     "pending",
		}

		err := repo.CreateExpression(expr1)
		require.NoError(t, err)
		err = repo.UpdateExpressionResult(expr1.ID, 10, "completed")
		require.NoError(t, err)

		err = repo.CreateExpression(expr2)
		require.NoError(t, err)

		// Получаем все выражения пользователя
		expressions, err := repo.GetExpressions(user.Username)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(expressions), 2)

		// Проверяем, что выражения есть в списке
		var found1, found2 bool
		for _, e := range expressions {
			if e.ID == expr1.ID {
				found1 = true
				assert.Equal(t, 10, e.Result)
				assert.Equal(t, "completed", e.Status)
			}
			if e.ID == expr2.ID {
				found2 = true
				assert.Equal(t, 0, e.Result)
				assert.Equal(t, "pending", e.Status)
			}
		}
		assert.True(t, found1)
		assert.True(t, found2)
	})

	// Тест для несуществующего выражения
	t.Run("NonExistentExpression", func(t *testing.T) {
		_, err := repo.GetExpressionByID(uuid.New())
		assert.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)
	})
}
