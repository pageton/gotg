package lua

import (
	"github.com/gotd/td/tg"
	"github.com/pageton/gotg/adapter"
	lua "github.com/yuin/gopher-lua"
)

const luaUpdateTypeName = "gotg.update"

func registerUpdateType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaUpdateTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), updateMethods))
}

var updateMethods = map[string]lua.LGFunction{
	"reply":         updateReply,
	"text":          updateText,
	"user_id":       updateUserID,
	"chat_id":       updateChatID,
	"msg_id":        updateMsgID,
	"first_name":    updateFirstName,
	"last_name":     updateLastName,
	"full_name":     updateFullName,
	"username":      updateUsername,
	"lang_code":     updateLangCode,
	"args":          updateArgs,
	"data":          updateData,
	"answer":        updateAnswerCallback,
	"delete":        updateDelete,
	"is_reply":      updateIsReply,
	"is_bot":        updateIsBot,
	"is_outgoing":   updateIsOutgoing,
	"is_incoming":   updateIsIncoming,
	"has_message":   updateHasMessage,
	"mention":       updateMention,
	"send_message":  updateSendMessage,
	"is_edited":     updateIsEdited,
	"connection_id": updateConnectionID,
	"is_business":   updateIsBusiness,
	"raw_call":      updateRawCall,
	"log_debug":     updateLogDebug,
	"log_info":      updateLogInfo,
	"log_success":   updateLogSuccess,
	"log_warn":      updateLogWarn,
	"log_error":     updateLogError,
}

func pushUpdate(L *lua.LState, u *adapter.Update) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = u
	L.SetMetatable(ud, L.GetTypeMetatable(luaUpdateTypeName))
	return ud
}

func checkUpdate(L *lua.LState) *adapter.Update {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*adapter.Update); ok {
		return v
	}
	L.ArgError(1, "gotg.update expected")
	return nil
}

func updateReply(L *lua.LState) int {
	u := checkUpdate(L)
	text := L.CheckString(2)
	_, err := u.Reply(text)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LTrue)
	return 1
}

func updateText(L *lua.LState) int {
	u := checkUpdate(L)
	L.Push(lua.LString(u.Text()))
	return 1
}

func updateUserID(L *lua.LState) int {
	u := checkUpdate(L)
	L.Push(lua.LNumber(u.UserID()))
	return 1
}

func updateChatID(L *lua.LState) int {
	u := checkUpdate(L)
	L.Push(lua.LNumber(u.ChatID()))
	return 1
}

func updateMsgID(L *lua.LState) int {
	u := checkUpdate(L)
	L.Push(lua.LNumber(u.MsgID()))
	return 1
}

func updateFirstName(L *lua.LState) int {
	u := checkUpdate(L)
	L.Push(lua.LString(u.FirstName()))
	return 1
}

func updateLastName(L *lua.LState) int {
	u := checkUpdate(L)
	L.Push(lua.LString(u.LastName()))
	return 1
}

func updateFullName(L *lua.LState) int {
	u := checkUpdate(L)
	L.Push(lua.LString(u.FullName()))
	return 1
}

func updateUsername(L *lua.LState) int {
	u := checkUpdate(L)
	L.Push(lua.LString(u.Username()))
	return 1
}

func updateLangCode(L *lua.LState) int {
	u := checkUpdate(L)
	L.Push(lua.LString(u.LangCode()))
	return 1
}

func updateArgs(L *lua.LState) int {
	u := checkUpdate(L)
	args := u.Args()
	tbl := L.NewTable()
	for i, arg := range args {
		tbl.RawSetInt(i+1, lua.LString(arg))
	}
	L.Push(tbl)
	return 1
}

func updateData(L *lua.LState) int {
	u := checkUpdate(L)
	L.Push(lua.LString(u.Data()))
	return 1
}

func updateAnswerCallback(L *lua.LState) int {
	u := checkUpdate(L)
	if u.CallbackQuery == nil {
		L.Push(lua.LNil)
		L.Push(lua.LString("not a callback query"))
		return 2
	}

	text := L.OptString(2, "")
	alert := false
	if L.GetTop() >= 3 {
		alert = L.OptBool(3, false)
	}

	_, err := u.Ctx.AnswerCallback(&tg.MessagesSetBotCallbackAnswerRequest{
		QueryID: u.CallbackQuery.QueryID,
		Message: text,
		Alert:   alert,
	})
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LTrue)
	return 1
}

func updateDelete(L *lua.LState) int {
	u := checkUpdate(L)
	err := u.Delete()
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LTrue)
	return 1
}

func updateIsReply(L *lua.LState) int {
	u := checkUpdate(L)
	L.Push(lua.LBool(u.IsReply()))
	return 1
}

func updateIsBot(L *lua.LState) int {
	u := checkUpdate(L)
	L.Push(lua.LBool(u.IsBot()))
	return 1
}

func updateIsOutgoing(L *lua.LState) int {
	u := checkUpdate(L)
	L.Push(lua.LBool(u.IsOutgoing()))
	return 1
}

func updateIsIncoming(L *lua.LState) int {
	u := checkUpdate(L)
	L.Push(lua.LBool(u.IsIncoming()))
	return 1
}

func updateHasMessage(L *lua.LState) int {
	u := checkUpdate(L)
	L.Push(lua.LBool(u.HasMessage()))
	return 1
}

func updateMention(L *lua.LState) int {
	u := checkUpdate(L)
	L.Push(lua.LString(u.Mention()))
	return 1
}

func updateSendMessage(L *lua.LState) int {
	u := checkUpdate(L)
	chatID := int64(L.CheckNumber(2))
	text := L.CheckString(3)
	_, err := u.SendMessage(chatID, text)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LTrue)
	return 1
}

func updateIsEdited(L *lua.LState) int {
	u := checkUpdate(L)
	L.Push(lua.LBool(u.IsEdited))
	return 1
}

func updateConnectionID(L *lua.LState) int {
	u := checkUpdate(L)
	L.Push(lua.LString(u.ConnectionID()))
	return 1
}

func updateIsBusiness(L *lua.LState) int {
	u := checkUpdate(L)
	L.Push(lua.LBool(u.IsBusinessUpdate()))
	return 1
}

func logArgs(L *lua.LState) (string, []any) {
	msg := L.CheckString(2)
	top := L.GetTop()
	var kvs []any
	for i := 3; i <= top; i++ {
		kvs = append(kvs, L.Get(i).String())
	}
	return msg, kvs
}

func updateLogDebug(L *lua.LState) int {
	u := checkUpdate(L)
	msg, kvs := logArgs(L)
	u.Log.Debug(msg, kvs...)
	return 0
}

func updateLogInfo(L *lua.LState) int {
	u := checkUpdate(L)
	msg, kvs := logArgs(L)
	u.Log.Info(msg, kvs...)
	return 0
}

func updateLogSuccess(L *lua.LState) int {
	u := checkUpdate(L)
	msg, kvs := logArgs(L)
	u.Log.Success(msg, kvs...)
	return 0
}

func updateLogWarn(L *lua.LState) int {
	u := checkUpdate(L)
	msg, kvs := logArgs(L)
	u.Log.Warn(msg, kvs...)
	return 0
}

func updateLogError(L *lua.LState) int {
	u := checkUpdate(L)
	msg, kvs := logArgs(L)
	u.Log.Error(msg, kvs...)
	return 0
}
