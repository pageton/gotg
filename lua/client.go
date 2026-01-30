package lua

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"github.com/pageton/gotg"
	"github.com/pageton/gotg/adapter"
	"github.com/pageton/gotg/dispatcher/handlers"
	"github.com/pageton/gotg/dispatcher/handlers/filters"
	gotglog "github.com/pageton/gotg/log"
	"github.com/pageton/gotg/session"
	lua "github.com/yuin/gopher-lua"
)

const luaClientTypeName = "gotg.client"

type luaClient struct {
	client *gotg.Client
	vm     *VM
}

func registerClientType(L *lua.LState) {
	mt := L.NewTypeMetatable(luaClientTypeName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), clientMethods))
}

var clientMethods = map[string]lua.LGFunction{
	"on_command":                  clientOnCommand,
	"on_message":                  clientOnMessage,
	"on_edited_message":           clientOnEditedMessage,
	"on_callback":                 clientOnCallback,
	"on_inline_query":             clientOnInlineQuery,
	"on_chosen_inline_result":     clientOnChosenInlineResult,
	"on_deleted_message":          clientOnDeletedMessage,
	"on_chat_member_updated":      clientOnChatMemberUpdated,
	"on_chat_join_request":        clientOnChatJoinRequest,
	"on_message_reaction":         clientOnMessageReaction,
	"on_chat_boost":               clientOnChatBoost,
	"on_outgoing":                 clientOnOutgoing,
	"on_update":                   clientOnUpdate,
	"on_business_message":         clientOnBusinessMessage,
	"on_business_edited_message":  clientOnBusinessEditedMessage,
	"on_business_deleted_message": clientOnBusinessDeletedMessage,
	"on_business_callback_query":  clientOnBusinessCallbackQuery,
	"on_business_connection":      clientOnBusinessConnection,
	"start":                       clientStart,
	"idle":                        clientIdle,
	"stop":                        clientStop,
}

func checkClient(L *lua.LState) *luaClient {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*luaClient); ok {
		return v
	}
	L.ArgError(1, "gotg.client expected")
	return nil
}

func luaNewClient(L *lua.LState) int {
	tbl := L.CheckTable(1)

	apiID := int(getNumberField(tbl, "api_id"))
	apiHash := getStringField(tbl, "api_hash")
	botToken := getStringField(tbl, "bot_token")
	phone := getStringField(tbl, "phone")
	inMemory := getBoolField(tbl, "in_memory")
	sessionPath := getStringField(tbl, "session")

	if botToken == "" && phone == "" {
		L.ArgError(1, "bot_token or phone required")
		return 0
	}
	if apiID == 0 {
		L.ArgError(1, "api_id is required")
		return 0
	}
	if apiHash == "" {
		L.ArgError(1, "api_hash is required")
		return 0
	}

	opts := &gotg.ClientOpts{
		InMemory:         inMemory,
		DisableCopyright: true,
	}

	if sessionPath != "" {
		opts.InMemory = false
	}

	logTbl := getTableField(tbl, "log")
	if logTbl != nil {
		cfg := &gotglog.Config{
			MinLevel:  gotglog.LevelInfo,
			Timestamp: true,
			Color:     true,
		}
		lvl := getStringField(logTbl, "level")
		switch lvl {
		case "debug":
			cfg.MinLevel = gotglog.LevelDebug
		case "warn":
			cfg.MinLevel = gotglog.LevelWarn
		case "error":
			cfg.MinLevel = gotglog.LevelError
		case "off":
			cfg.MinLevel = gotglog.LevelOff
		}
		if v := getOptBoolField(logTbl, "color"); v != nil {
			cfg.Color = *v
		}
		if v := getOptBoolField(logTbl, "timestamp"); v != nil {
			cfg.Timestamp = *v
		}
		logFile := getStringField(logTbl, "file")
		if logFile != "" {
			cfg.LogFile = logFile
		}
		opts.LogConfig = cfg
	}

	if sessionPath != "" {
		opts.Session = session.SqlSession(sqlite.Open(sessionPath))
	}

	var client *gotg.Client
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r)
			}
		}()
		if botToken != "" {
			client, err = gotg.NewClient(apiID, apiHash, gotg.AsBot(botToken), opts)
		} else {
			client, err = gotg.NewClient(apiID, apiHash, gotg.AsUser(phone), opts)
		}
	}()
	if err != nil {
		L.RaiseError("failed to create client: %s", err.Error())
		return 0
	}

	registry := L.Get(lua.RegistryIndex)
	var vm *VM
	if tbl, ok := registry.(*lua.LTable); ok {
		if v := tbl.RawGetString("__gotg_vm"); v != lua.LNil {
			if ud, ok := v.(*lua.LUserData); ok {
				vm = ud.Value.(*VM)
			}
		}
	}

	lc := &luaClient{
		client: client,
		vm:     vm,
	}

	ud := L.NewUserData()
	ud.Value = lc
	L.SetMetatable(ud, L.GetTypeMetatable(luaClientTypeName))
	L.Push(ud)
	return 1
}

func clientOnCommand(L *lua.LState) int {
	lc := checkClient(L)
	name := L.CheckString(2)
	fn := L.CheckFunction(3)

	var updateFilters []filters.UpdateFilter
	if L.GetTop() >= 4 {
		filterStr := L.CheckString(4)
		if f := resolveUpdateFilter(filterStr); f != nil {
			updateFilters = append(updateFilters, f)
		}
	}

	handler := makeUpdateHandler(lc, fn)
	lc.client.Dispatcher.AddHandler(handlers.OnCommand(name, handler, updateFilters...))
	return 0
}

func clientOnMessage(L *lua.LState) int {
	lc := checkClient(L)
	fn := L.CheckFunction(2)
	msgFilters := parseMessageFilters(L, 3)
	handler := makeUpdateHandler(lc, fn)
	lc.client.Dispatcher.AddHandler(handlers.OnMessage(handler, msgFilters...))
	return 0
}

func parseMessageFilters(L *lua.LState, argPos int) []filters.MessageFilter {
	var result []filters.MessageFilter
	if L.GetTop() < argPos {
		return result
	}
	filterArg := L.Get(argPos)
	switch v := filterArg.(type) {
	case lua.LString:
		if f := resolveMessageFilter(string(v)); f != nil {
			result = append(result, f)
		}
	case *lua.LTable:
		filterStr := getStringField(v, "filter")
		if filterStr != "" {
			if f := resolveMessageFilter(filterStr); f != nil {
				result = append(result, f)
			}
		}
	}
	return result
}

func clientOnCallback(L *lua.LState) int {
	lc := checkClient(L)

	var filter filters.CallbackQueryFilter
	var fn *lua.LFunction

	switch L.Get(2).Type() {
	case lua.LTString:
		prefix := L.CheckString(2)
		fn = L.CheckFunction(3)
		filter = filters.CallbackQuery.Prefix(prefix)
	case lua.LTFunction:
		fn = L.CheckFunction(2)
	default:
		L.ArgError(2, "string or function expected")
		return 0
	}

	handler := makeUpdateHandler(lc, fn)
	lc.client.Dispatcher.AddHandler(handlers.OnCallbackQuery(filter, handler))
	return 0
}

func clientOnEditedMessage(L *lua.LState) int {
	lc := checkClient(L)
	fn := L.CheckFunction(2)
	msgFilters := parseMessageFilters(L, 3)
	handler := makeUpdateHandler(lc, fn)
	lc.client.Dispatcher.AddHandler(handlers.OnEditedMessage(handler, msgFilters...))
	return 0
}

func clientOnInlineQuery(L *lua.LState) int {
	lc := checkClient(L)
	fn := L.CheckFunction(2)
	handler := makeUpdateHandler(lc, fn)
	lc.client.Dispatcher.AddHandler(handlers.OnInlineQuery(handler))
	return 0
}

func clientOnChosenInlineResult(L *lua.LState) int {
	lc := checkClient(L)
	fn := L.CheckFunction(2)
	handler := makeUpdateHandler(lc, fn)
	lc.client.Dispatcher.AddHandler(handlers.OnChosenInlineResult(handler))
	return 0
}

func clientOnDeletedMessage(L *lua.LState) int {
	lc := checkClient(L)
	fn := L.CheckFunction(2)
	handler := makeUpdateHandler(lc, fn)
	lc.client.Dispatcher.AddHandler(handlers.OnDeletedMessage(handler))
	return 0
}

func clientOnChatMemberUpdated(L *lua.LState) int {
	lc := checkClient(L)
	fn := L.CheckFunction(2)
	handler := makeUpdateHandler(lc, fn)
	lc.client.Dispatcher.AddHandler(handlers.OnChatMemberUpdated(handler))
	return 0
}

func clientOnChatJoinRequest(L *lua.LState) int {
	lc := checkClient(L)
	fn := L.CheckFunction(2)
	handler := makeUpdateHandler(lc, fn)
	lc.client.Dispatcher.AddHandler(handlers.OnChatJoinRequest(handler))
	return 0
}

func clientOnMessageReaction(L *lua.LState) int {
	lc := checkClient(L)
	fn := L.CheckFunction(2)
	handler := makeUpdateHandler(lc, fn)
	lc.client.Dispatcher.AddHandler(handlers.OnMessageReaction(handler))
	return 0
}

func clientOnChatBoost(L *lua.LState) int {
	lc := checkClient(L)
	fn := L.CheckFunction(2)
	handler := makeUpdateHandler(lc, fn)
	lc.client.Dispatcher.AddHandler(handlers.OnChatBoost(handler))
	return 0
}

func clientOnOutgoing(L *lua.LState) int {
	lc := checkClient(L)
	fn := L.CheckFunction(2)
	msgFilters := parseMessageFilters(L, 3)
	handler := makeUpdateHandler(lc, fn)
	lc.client.Dispatcher.AddHandler(handlers.OnOutgoing(handler, msgFilters...))
	return 0
}

func clientOnUpdate(L *lua.LState) int {
	lc := checkClient(L)
	fn := L.CheckFunction(2)
	handler := makeUpdateHandler(lc, fn)
	lc.client.Dispatcher.AddHandler(handlers.OnUpdate(handler))
	return 0
}

func clientOnBusinessMessage(L *lua.LState) int {
	lc := checkClient(L)
	fn := L.CheckFunction(2)
	msgFilters := parseMessageFilters(L, 3)
	handler := makeUpdateHandler(lc, fn)
	lc.client.Dispatcher.AddHandler(handlers.OnBusinessMessage(handler, msgFilters...))
	return 0
}

func clientOnBusinessEditedMessage(L *lua.LState) int {
	lc := checkClient(L)
	fn := L.CheckFunction(2)
	msgFilters := parseMessageFilters(L, 3)
	handler := makeUpdateHandler(lc, fn)
	lc.client.Dispatcher.AddHandler(handlers.OnBusinessEditedMessage(handler, msgFilters...))
	return 0
}

func clientOnBusinessDeletedMessage(L *lua.LState) int {
	lc := checkClient(L)
	fn := L.CheckFunction(2)
	handler := makeUpdateHandler(lc, fn)
	lc.client.Dispatcher.AddHandler(handlers.OnBusinessDeletedMessage(handler))
	return 0
}

func clientOnBusinessCallbackQuery(L *lua.LState) int {
	lc := checkClient(L)
	fn := L.CheckFunction(2)
	handler := makeUpdateHandler(lc, fn)
	lc.client.Dispatcher.AddHandler(handlers.OnBusinessCallbackQuery(handler))
	return 0
}

func clientOnBusinessConnection(L *lua.LState) int {
	lc := checkClient(L)
	fn := L.CheckFunction(2)
	handler := makeUpdateHandler(lc, fn)
	lc.client.Dispatcher.AddHandler(handlers.OnBusinessConnection(handler))
	return 0
}

func clientStart(L *lua.LState) int {
	// client is already started by NewClient; this is a no-op kept for API symmetry
	return 0
}

func clientIdle(L *lua.LState) int {
	lc := checkClient(L)
	err := lc.client.Idle()
	if err != nil {
		L.RaiseError("client error: %s", err.Error())
	}
	return 0
}

func clientStop(L *lua.LState) int {
	lc := checkClient(L)
	lc.client.Stop()
	return 0
}

func makeUpdateHandler(lc *luaClient, fn *lua.LFunction) func(*adapter.Update) error {
	return func(u *adapter.Update) error {
		if lc.vm == nil {
			return nil
		}
		luaUpdate := pushUpdate(lc.vm.L, u)
		return lc.vm.CallLuaFunc(fn, luaUpdate)
	}
}
