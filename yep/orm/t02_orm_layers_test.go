// Copyright 2016 NDP Systèmes. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package orm

import (
	"testing"
	"time"
)

type User_Full struct {
	ID           int
	UserName     string
	Email        string
	Password     string
	Status       int16
	IsStaff      bool
	IsActive     bool
	Created      time.Time
	Updated      time.Time
	Profile      *Profile
	Posts        []*Post
	ShouldSkip   string
	Nums         int
	Langs        SliceStringField
	Extra        JSONField
	unexport     bool
	unexportBool bool
	Email2       string
	IsPremium    bool
}

type Tag_Full struct {
	ID          int
	Name        string
	Description string
	BestPost    *Post
	Posts       []*Post
}

type Tag_Partial struct {
	ID          int
	Name        string
	Description string
	Posts       []*Post
}

type User_Partial struct {
	ID        int
	Email     string
	Email2    string
	IsPremium bool
	Profile   *Profile_Partial
	Created   time.Time
	Updated   time.Time
	Extra     JSONField
	UserName  string
}

type Profile_Partial struct {
	ID      int64
	Age     int16
	Country string
}

type User_PartialWithPosts struct {
	ID        int
	Email     string
	Email2    string
	IsPremium bool
	Profile   *Profile_PartialWithBestPost
	Posts     []*Post
}

type Profile_PartialWithBestPost struct {
	ID       int64
	Age      int16
	Country  string
	BestPost *Post
}

var (
	userJohnID int64
	userJaneID int64
	err        error
)

func TestFullUser(t *testing.T) {
	// Insert full users
	user1 := User_Full{
		UserName: "John Smith",
		Email:    "jsmith@example.com",
		Status:   132,
		IsActive: true,
	}
	userJohnID, err = dORM.Insert(&user1)
	throwFail(t, err)

	user2 := User_Full{
		UserName: "Jane Smith",
		Email:    "j2smith@example.com",
		Email2:   "jane.smith@example.net",
		Status:   12,
		IsActive: true,
	}
	userJaneID, err = dORM.Insert(&user2)
	throwFail(t, err)

	// Query full user
	user := User_Full{ID: int(userJohnID)}
	err = dORM.Read(&user)
	throwFail(t, err)
	throwFail(t, AssertIs(user.ID, userJohnID))
	throwFail(t, AssertIs(user.IsStaff, false))
	throwFail(t, AssertIs(user.UserName, "John Smith"))
	throwFail(t, AssertIs(user.Email2, ""))
	throwFail(t, AssertIs(user.IsPremium, false))
	throwFail(t, AssertIs(user.IsActive, true))
	throwFail(t, AssertIs(user.Status, 132))

	// Query through QuerySet with full user
	var user3 User_Full
	qs := dORM.QueryTable("user")
	err = qs.Filter("email2", "jane.smith@example.net").One(&user3)
	throwFail(t, err)
	throwFail(t, AssertIs(user3.ID, userJaneID))
	throwFail(t, AssertIs(user3.Email, "j2smith@example.com"))

	// Update with full user
	user3.Email = "jsmith2@example.com"
	user3.IsPremium = true
	num, err := dORM.Update(&user3)
	throwFail(t, err)
	throwFail(t, AssertIs(num, 1))
}

func TestPartialUser(t *testing.T) {
	user3 := User_Full{
		UserName: "Will Smith",
		Email:    "wsmith@example.com",
		Email2:   "will.smith@example.net",
		Status:   12,
		IsActive: true,
	}
	id3, err := dORM.Insert(&user3)
	throwFail(t, err)

	// Query partial user
	user := User_Partial{ID: int(id3)}
	err = dORM.Read(&user)
	throwFail(t, err)
	throwFail(t, AssertIs(user.ID, id3))
	throwFail(t, AssertIs(user.Email, "wsmith@example.com"))
	throwFail(t, AssertIs(user.Email2, "will.smith@example.net"))
	throwFail(t, AssertIs(user.IsPremium, false))

	// Query through queryset
	var users []*User_Partial
	qs := dORM.QueryTable(new(User_Partial))
	num, err := qs.Filter("username__contains", "Smith").All(&users)
	throwFail(t, err)
	throwFail(t, AssertIs(num, 3))

	// Update with partial user
	user.IsPremium = true
	num, err = dORM.Update(&user)
	throwFail(t, err)
	throwFail(t, AssertIs(num, 1))

	// Re-read with full user and check update
	userFull := User_Full{ID: int(id3)}
	err = dORM.Read(&userFull)
	throwFail(t, err)
	throwFail(t, AssertIs(userFull.ID, id3))
	throwFail(t, AssertIs(userFull.Email, "wsmith@example.com"))
	throwFail(t, AssertIs(userFull.IsPremium, true))

	// Delete from partial
	num, err = dORM.Delete(&user)
	throwFail(t, err)
	throwFail(t, AssertIs(num, 1))
}

func TestRelPartial(t *testing.T) {
	// One2one
	user_jane := User{ID: int(userJaneID)}
	err = dORM.Read(&user_jane)
	throwFail(t, err)

	user_jane.Profile = &Profile{
		Age:   24,
		Money: 1234,
	}
	_, err := dORM.Insert(user_jane.Profile)
	throwFail(t, err)
	num, err := dORM.Update(&user_jane)
	throwFail(t, err)
	throwFail(t, AssertIs(num, 1))

	// Read
	user := User_Partial{ID: int(userJaneID)}
	err = dORM.Read(&user)
	throwFail(t, err)
	err = dORM.Read(user.Profile)
	throwFail(t, err)
	throwFail(t, AssertIs(user.Profile.Age, 24))
	throwFail(t, AssertIs(user.Profile.Country, ""))
	user.Profile.Country = "UK"
	num, err = dORM.Update(user.Profile)
	throwFail(t, err)
	throwFail(t, AssertIs(num, 1))

	// Read with query set
	var user2 User_Partial
	err = dORM.QueryTable("User").Filter("UserName", "Jane Smith").RelatedSel().One(&user2)
	throwFail(t, err)
	throwFail(t, AssertIs(user2.ID, userJaneID))
	throwFail(t, AssertIs(user2.Email, "jsmith2@example.com"))
	throwFail(t, AssertIs(user2.Profile.Age, 24))
	throwFail(t, AssertIs(user2.Profile.Country, "UK"))

	// Test One2many
	post1 := Post{
		User:    &user_jane,
		Title:   "Title of post1",
		Content: "Content of post1",
	}
	dORM.Insert(&post1)
	post2 := Post{
		User:    &user_jane,
		Title:   "Title of post2",
		Content: "Content of post2",
	}
	dORM.Insert(&post2)
	// Loadrelated with query set on full user
	var user_full User_Full
	err = dORM.QueryTable(new(User_Full)).Filter("UserName", "Jane Smith").One(&user_full)
	throwFail(t, err)
	num, err = dORM.LoadRelated(&user_full, "Posts")
	throwFail(t, err)
	throwFail(t, AssertIs(num, 2))
	throwFail(t, AssertIs(len(user_full.Posts), 2))
	throwFail(t, AssertIs(user_full.Posts[0].Title, "Title of post1"))

	// Loadrelated with query set on partial user
	var user_partial User_PartialWithPosts
	err = dORM.QueryTable(new(User_Full)).Filter("UserName", "Jane Smith").One(&user_partial)
	throwFail(t, err)
	num, err = dORM.LoadRelated(&user_partial, "Posts")
	throwFail(t, err)
	throwFail(t, AssertIs(num, 2))
	throwFail(t, AssertIs(len(user_partial.Posts), 2))
	throwFail(t, AssertIs(user_partial.Posts[0].Title, "Title of post1"))

	// Extra related tests
	user_jane.Profile.BestPost = &post1
	dORM.Update(user_jane.Profile)

	var user3 User_PartialWithPosts
	err = dORM.QueryTable("User").Filter("UserName", "Jane Smith").RelatedSel().One(&user3)
	throwFail(t, err)
	throwFail(t, AssertIs(user3.Profile.Country, "UK"))
	throwFail(t, AssertIs(user3.Profile.BestPost.ID, post1.ID))
}

func TestM2MPartial(t *testing.T) {
	// Get Jane's posts
	user_jane := User{ID: int(userJaneID)}
	err = dORM.Read(&user_jane)
	throwFail(t, err)
	_, err = dORM.LoadRelated(&user_jane, "Posts")
	throwFail(t, err)

	// Add first 3 tags to Jane's first post
	var tags []*Tag_Partial
	_, err = dORM.QueryTable("Tag").Limit(3).All(&tags)
	throwFail(t, err)
	tags[0].Description = "Tag0 description"
	_, err = dORM.Update(tags[0])
	throwFail(t, err)
	post1Tags := dORM.QueryM2M(user_jane.Posts[0], "Tags")
	post1Tags.Add(tags)

	// Get Tags to check
	num, err := dORM.LoadRelated(user_jane.Posts[0], "Tags")
	throwFail(t, err)
	throwFail(t, AssertIs(num, 3))
	throwFail(t, AssertIs(len(user_jane.Posts[0].Tags), 3))
}

func TestMulti(t *testing.T) {
	// Insert multi
	user1 := &User_Partial{
		UserName: "User 1",
		Email:    "user1@example.com",
	}
	user2 := &User_Partial{
		UserName: "User 2",
		Email:    "user2@example.com",
		Email2:   "userthesecond@example.net",
	}
	user3 := &User_Partial{
		UserName:  "User 3",
		Email:     "user3@example.com",
		IsPremium: true,
	}
	_, err = dORM.InsertMulti(3, []*User_Partial{user1, user2, user3})
	throwFail(t, err)

	// Update
	num, err := dORM.QueryTable(user1).Filter("Email__contains", "user").Update(Params{"status": 14})
	throwFail(t, err)
	throwFail(t, AssertIs(num, 3))

	// Query
	var users []*User_Partial
	num, err = dORM.QueryTable("user").Filter("Email__contains", "user").OrderBy("UserName").All(&users)
	throwFail(t, err)
	throwFail(t, AssertIs(num, 3))
	throwFail(t, AssertIs(users[0].Email, "user1@example.com"))
	throwFail(t, AssertIs(users[1].Email, "user2@example.com"))
	throwFail(t, AssertIs(users[1].Email2, "userthesecond@example.net"))
	throwFail(t, AssertIs(users[2].Email, "user3@example.com"))
}
