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
	Language   string `json:"language"`
	TimeZone   string `json:"timezone"`
}

type ChangePwdData struct {
	Username    string `json:"username"`
	OldPassword string `json:"oldpassword"`
	NewPassword string `json:"newpassword"`
}

var TableName string = "users"
var LoginQuery string = `SELECT users.id as ID,users.name as Name,users.lastname as LastName, IFNULL(users.languageid,0) as languageid, IFNULL(users.timezoneid,0) as timezoneid, IFNULL(languages.name,'') as LanguageCode, IFNULL(timezones.name,'') as TimeZoneCode, Password FROM users 
							left join languages on languages.id = users.languageid
							left join timezones on timezones.id = users.timezoneid
							WHERE LoginName= '%s'`

//AND (Password='%s' OR Password is null OR Password='')"
var GetUserImageQuery string = "SELECT pictureurl as PictureUrl, loginname as LoginName FROM users WHERE loginname='%s'"

/*`Select ID, Name, Description, LngCode, PageType, ParentID, Page, Inputs, Icon,
CASE WHEN Exists (SELECT 1 FROM menu_roles MR INNER JOIN user_roles UR ON UR.RoleID = MR.RoleID
WHERE MenuID = M.ID AND UserID = %d AND COALESCE(ViewOnly,0) =0) THEN 0 ELSE 1 END As ViewOnly, Position
FROM menus  M
WHERE COALESCE(Mobile,0) = %d AND COALESCE(Desktop,0) = %d AND COALESCE(ParentID,0) = %d
	AND COALESCE(MenuShow,0) = 1
	AND EXISTS
	(SELECT 1 FROM menu_roles MR
		INNER JOIN user_roles UR ON UR.RoleID = MR.RoleID
		WHERE UR.UserID = %d AND MR.menuid = M.ID)
Order By Position `  */
var GetUserMenusQuery string = `Select M.id as ID, M.name as Name, COALESCE(lc.mediumtext_, shorttext,M.name) as Description,  COALESCE(lnc.name, '') as  LngCode, pagetype as PageType, parentid as ParentID, CASE WHEN pagetype = 3 THEN url ELSE page END as Page, inputs as Inputs, icon as Icon, M.url as Url,
	 CASE WHEN Exists (SELECT 1 FROM menu_roles MR INNER JOIN user_roles UR ON UR.roleid = MR.roleid 
					WHERE menuid = M.id AND userid = U.id AND COALESCE(viewonly,0) =0) THEN 0 
		ELSE 1 END As ViewOnly, position  AS Position
FROM menus  M 
	inner join users U ON  1=1
	left join lngcodes lnc on lnc.id = M.lngcodeid
	left join lngcode_contents lc on lc.lngcodeid = lnc.id and ( lc.languageid = U.languageid OR
		(lc.languageid = 1 and not exists (select 1 FROM lngcode_contents WHERE languageid = languageid and lngcodeid = M.lngcodeid)))
WHERE U.id = %d
	AND ( (COALESCE(mobile,0) = 1 AND 1 = %d ) OR ( COALESCE(desktop,0) = 1  AND 0 = %d) )
	AND COALESCE(parentid,0) = %d
	AND COALESCE(menushow,0) = 1 
	AND EXISTS (SELECT 1 FROM menu_roles MR 
				INNER JOIN user_roles UR ON UR.roleid = MR.roleid 
				WHERE UR.userid =U.id AND MR.menuid = M.id) 
Order By position`
