package mariadb

import (
	"context"
	"testing"
	"time"

	"github.com/roledio/roled/auth/internal/entities"
	"github.com/roledio/roled/auth/internal/models"
	"github.com/roledio/roled/auth/internal/repositories/mariadb/testutil"
	pm "github.com/roledio/roled/auth/pkg/models"
	"github.com/stretchr/testify/require"
)

func TestMemberQueriesAndSoftDeletion(t *testing.T) {
	ctx := context.Background()
	db := testSuite.GetDB()
	repo := NewMemberRepository(db)
	testSuite.CleanTables(t, "members", "users", "projects", "accounts")
	for _, id := range []string{"member-owner", "member-other"} {
		_, err := testutil.CreateAccount(ctx, db, testutil.AccountFixture{ID: id, Name: id, IsActive: true})
		require.NoError(t, err)
	}
	_, err := testutil.CreateProject(ctx, db, testutil.ProjectFixture{ID: "member-project", AccountID: "member-owner", Name: "Project", IsActive: true})
	require.NoError(t, err)
	stamp := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	for i, name := range []string{"Alice", "Bob", "Foreign"} {
		email := name + "@example.com"
		_, err = testutil.CreateUser(ctx, db, testutil.UserFixture{ID: name, AccountID: "member-owner", ProjectID: "member-project", DisplayName: name, Email: &email, IsActive: i != 1})
		require.NoError(t, err)
		account := "member-owner"
		if i == 2 {
			account = "member-other"
		}
		require.NoError(t, repo.Create(ctx, &entities.Member{ID: "member-" + name, AccountID: account, UserID: name, IsAdmin: i == 0}))
		_, err = db.ExecContext(ctx, "UPDATE members SET created_at = ? WHERE id = ?", stamp.Add(time.Duration(i)*time.Hour), "member-"+name)
		require.NoError(t, err)
	}
	_, err = db.ExecContext(ctx, "UPDATE users SET email_verified_at = ? WHERE id = ?", stamp, "Alice")
	require.NoError(t, err)
	yes, no := true, false
	until := stamp
	since := stamp.Add(time.Hour)
	for _, tc := range []struct {
		name string
		req  models.GetMembersRequest
		want []string
	}{
		{"all", models.GetMembersRequest{}, []string{"member-Alice", "member-Bob"}},
		{"active", models.GetMembersRequest{IsActive: &yes}, []string{"member-Alice"}},
		{"inactive", models.GetMembersRequest{IsActive: &no}, []string{"member-Bob"}},
		{"verified", models.GetMembersRequest{IsVerified: &yes}, []string{"member-Alice"}},
		{"unverified", models.GetMembersRequest{IsVerified: &no}, []string{"member-Bob"}},
		{"admin", models.GetMembersRequest{IsAdmin: &yes}, []string{"member-Alice"}},
		{"nonadmin", models.GetMembersRequest{IsAdmin: &no}, []string{"member-Bob"}},
		{"name", models.GetMembersRequest{Search: " Alice "}, []string{"member-Alice"}},
		{"email", models.GetMembersRequest{Search: "Bob@example"}, []string{"member-Bob"}},
		{"until inclusive", models.GetMembersRequest{CreatedAtUntil: &until}, []string{"member-Alice"}},
		{"since inclusive", models.GetMembersRequest{CreatedAtSince: &since}, []string{"member-Bob"}},
		{"combined", models.GetMembersRequest{IsAdmin: &yes, IsActive: &no}, []string{}},
		{"injection", models.GetMembersRequest{Search: "' OR 1=1 --"}, []string{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.req.AccountID = "member-owner"
			tc.req.SortBy = "name"
			list, e := repo.FindAll(ctx, &tc.req)
			require.NoError(t, e)
			ids := []string{}
			for _, v := range list {
				ids = append(ids, v.ID)
			}
			require.Equal(t, tc.want, ids)
			n, e := repo.Count(ctx, &tc.req)
			require.NoError(t, e)
			require.Equal(t, len(tc.want), n)
		})
	}
	for _, sort := range []string{"name", "is_active", "is_admin", "updated_at", "created_at", "name; DROP TABLE members"} {
		list, e := repo.FindAll(ctx, &models.GetMembersRequest{AccountID: "member-owner", PageRequest: pm.PageRequest{SortBy: sort, SortDir: "desc", PageSize: 1, PageNum: 2}})
		require.NoError(t, e)
		require.Len(t, list, 1)
		if sort == "name" {
			require.Equal(t, "member-Alice", list[0].ID)
		}
	}
	for _, filter := range []*bool{nil, &yes, &no} {
		n, e := repo.CountByAccountID(ctx, "member-owner", filter)
		require.NoError(t, e)
		if filter == nil {
			require.Equal(t, 2, n)
		} else {
			require.Equal(t, 1, n)
		}
	}
	missing, e := repo.FindByID(ctx, "missing")
	require.NoError(t, e)
	require.Nil(t, missing)
	missing, e = repo.FindByAccountIDAndUserID(ctx, "member-other", "Alice")
	require.NoError(t, e)
	require.Nil(t, missing)
	missingJoined, e := repo.FindByIDJoinUser(ctx, "missing")
	require.NoError(t, e)
	require.Nil(t, missingJoined)
	alice, e := repo.FindByAccountIDAndUserID(ctx, "member-owner", "Alice")
	require.NoError(t, e)
	require.Equal(t, "member-Alice", alice.ID)
	joined, e := repo.FindByIDJoinUser(ctx, alice.ID)
	require.NoError(t, e)
	require.Equal(t, "Alice", joined.DisplayName)
	require.True(t, joined.IsVerified)
	require.True(t, joined.IsActive)
	alice.IsAdmin = false
	n, e := repo.Update(ctx, alice)
	require.NoError(t, e)
	require.Equal(t, 1, n)
	updated, e := repo.FindByID(ctx, alice.ID)
	require.NoError(t, e)
	require.False(t, updated.IsAdmin)
	n, e = repo.Delete(ctx, alice)
	require.NoError(t, e)
	require.Equal(t, 1, n)
	updated, e = repo.FindByID(ctx, alice.ID)
	require.NoError(t, e)
	require.Nil(t, updated)
	updated, e = repo.FindByAccountIDAndUserID(ctx, "member-owner", "Alice")
	require.NoError(t, e)
	require.Nil(t, updated)
	joined, e = repo.FindByIDJoinUser(ctx, alice.ID)
	require.NoError(t, e)
	require.Nil(t, joined)
	n, e = repo.Update(ctx, alice)
	require.NoError(t, e)
	require.Zero(t, n)
	n, e = repo.Delete(ctx, alice)
	require.NoError(t, e)
	require.Zero(t, n)
	n, e = repo.Count(ctx, &models.GetMembersRequest{AccountID: "member-owner"})
	require.NoError(t, e)
	require.Equal(t, 1, n)
	n, e = repo.DeleteByAccountID(ctx, "member-owner")
	require.NoError(t, e)
	require.Equal(t, 1, n)
	n, e = repo.CountByAccountID(ctx, "member-owner", nil)
	require.NoError(t, e)
	require.Zero(t, n)
	foreign, e := repo.FindByID(ctx, "member-Foreign")
	require.NoError(t, e)
	require.NotNil(t, foreign)
	var physicallyStored int
	require.NoError(t, db.GetContext(ctx, &physicallyStored, "SELECT COUNT(*) FROM members WHERE account_id = ? AND deleted_at IS NOT NULL", "member-owner"))
	require.Equal(t, 2, physicallyStored)
}

func TestMemberRepositoryCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := NewMemberRepository(testSuite.GetDB())
	for _, tc := range []struct {
		name string
		run  func() error
	}{
		{"create", func() error { return r.Create(ctx, &entities.Member{ID: "cancel"}) }},
		{"list", func() error { _, e := r.FindAll(ctx, &models.GetMembersRequest{AccountID: "a"}); return e }},
		{"count", func() error { _, e := r.Count(ctx, &models.GetMembersRequest{AccountID: "a"}); return e }},
		{"count account", func() error { _, e := r.CountByAccountID(ctx, "a", nil); return e }},
		{"find account user", func() error { _, e := r.FindByAccountIDAndUserID(ctx, "a", "u"); return e }},
		{"find id", func() error { _, e := r.FindByID(ctx, "m"); return e }},
		{"find joined", func() error { _, e := r.FindByIDJoinUser(ctx, "m"); return e }},
		{"update", func() error { _, e := r.Update(ctx, &entities.Member{ID: "m"}); return e }},
		{"delete", func() error { _, e := r.Delete(ctx, &entities.Member{ID: "m"}); return e }},
		{"delete account", func() error { _, e := r.DeleteByAccountID(ctx, "a"); return e }},
	} {
		t.Run(tc.name, func(t *testing.T) { require.ErrorIs(t, tc.run(), context.Canceled) })
	}
}
