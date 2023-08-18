// Copyright 2023 IAC. All Rights Reserved.
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

package user

type LoginUserData struct {
	ID       int    `json:"id"` // The user's unique ID
	Username string `json:"username"`
	Password string `json:"password"`
	ClientID string `json:"clientid"`
	Token    string `json:"token"`
	Renew    bool   `json:"renew"`
}

type User struct {
	ID         int    `json:"id"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	ClientID   string `json:"clientid"`
	Token      string `json:"token"`
	CreatedOn  string `json:"createdon"`
	ExpirateOn string `json:"expirateon"`
	Email      string `json:"email"`
}

type ChangePwdData struct {
	Username    string `json:"username"`
	OldPassword string `json:"oldpassword"`
	NewPassword string `json:"newpassword"`
}

var TableName string = "users"
var LoginQuery string = "SELECT ID,Name,LastName, LanguageCode, TimeZoneCode, Password FROM users WHERE LoginName='%s' "

//AND (Password='%s' OR Password is null OR Password='')"
var GetUserImageQuery string = "SELECT PictureUrl, LoginName FROM users WHERE LoginName='%s'"

var GetUserMenusQuery string = "Select ID, Name, Description, LngCode, PageType, ParentID, Page, Inputs, Icon, CASE WHEN Exists (SELECT 1 FROM MENU_ROLES MR INNER JOIN USER_ROLES UR ON UR.RoleID = MR.RoleID WHERE MenuID = M.ID AND UserID = %d AND COALESCE(ViewOnly,0) =0) THEN 0 ELSE 1 END As ViewOnly FROM MENUS  M WHERE COALESCE(Mobile,0) = %d AND COALESCE(Desktop,0) = %d AND COALESCE(ParentID,0) = %d AND COALESCE(MenuShow,0) = 1 AND EXISTS (    SELECT 1 FROM MENU_ROLES MR INNER JOIN USER_ROLES UR ON UR.RoleID = MR.RoleID WHERE UR.UserID = %d AND MR.menuid = M.ID) "
