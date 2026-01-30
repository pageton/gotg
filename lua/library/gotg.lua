---@meta gotg

local gotg = {}

---@class gotg.ClientConfig
---@field api_id number|nil
---@field api_hash string|nil
---@field bot_token? string
---@field phone? string
---@field in_memory? boolean
---@field session? string
---@field log? gotg.LogConfig

---@class gotg.LogConfig
---@field level? "debug"|"info"|"warn"|"error"|"off"
---@field color? boolean
---@field timestamp? boolean
---@field file? string

---@param config gotg.ClientConfig
---@return gotg.Client
function gotg.new_client(config) end

---@class gotg.Client
local Client = {}

---@alias gotg.UpdateHandler fun(u: gotg.Update)

---@alias gotg.UpdateFilter
---| "private"
---| "group"
---| "supergroup"
---| "channel"
---| "incoming"
---| "outgoing"
---| "business"

---@alias gotg.MessageFilter string
---| "text"
---| "photo"
---| "video"
---| "audio"
---| "voice"
---| "sticker"
---| "animation"
---| "document"
---| "media"
---| "reply"
---| "forwarded"
---| "edited"
---| "poll"
---| "dice"
---| "game"
---| "contact"
---| "location"
---| "venue"
---| "video_note"
---| "web_page"
---| "private"
---| "group"
---| "channel"
---| "incoming"
---| "outgoing"
---| "command"

---@param name string
---@param handler gotg.UpdateHandler
---@param filter? gotg.UpdateFilter
function Client:on_command(name, handler, filter) end

---@param handler gotg.UpdateHandler
---@param filter? gotg.MessageFilter
function Client:on_message(handler, filter) end

---@param handler gotg.UpdateHandler
---@param filter? gotg.MessageFilter
function Client:on_edited_message(handler, filter) end

---@overload fun(self: gotg.Client, handler: gotg.UpdateHandler)
---@overload fun(self: gotg.Client, prefix: string, handler: gotg.UpdateHandler)
function Client:on_callback(...) end

---@param handler gotg.UpdateHandler
function Client:on_inline_query(handler) end

---@param handler gotg.UpdateHandler
function Client:on_chosen_inline_result(handler) end

---@param handler gotg.UpdateHandler
function Client:on_deleted_message(handler) end

---@param handler gotg.UpdateHandler
function Client:on_chat_member_updated(handler) end

---@param handler gotg.UpdateHandler
function Client:on_chat_join_request(handler) end

---@param handler gotg.UpdateHandler
function Client:on_message_reaction(handler) end

---@param handler gotg.UpdateHandler
function Client:on_chat_boost(handler) end

---@param handler gotg.UpdateHandler
---@param filter? gotg.MessageFilter
function Client:on_outgoing(handler, filter) end

---@param handler gotg.UpdateHandler
function Client:on_update(handler) end

---@param handler gotg.UpdateHandler
---@param filter? gotg.MessageFilter
function Client:on_business_message(handler, filter) end

---@param handler gotg.UpdateHandler
---@param filter? gotg.MessageFilter
function Client:on_business_edited_message(handler, filter) end

---@param handler gotg.UpdateHandler
function Client:on_business_deleted_message(handler) end

---@param handler gotg.UpdateHandler
function Client:on_business_callback_query(handler) end

---@param handler gotg.UpdateHandler
function Client:on_business_connection(handler) end

function Client:start() end

function Client:idle() end

function Client:stop() end

---@class gotg.Update
local Update = {}

---@param text string
---@return boolean|nil ok
---@return string? err
function Update:reply(text) end

---@return string
function Update:text() end

---@return number
function Update:user_id() end

---@return number
function Update:chat_id() end

---@return number
function Update:msg_id() end

---@return string
function Update:first_name() end

---@return string
function Update:last_name() end

---@return string
function Update:full_name() end

---@return string
function Update:username() end

---@return string
function Update:lang_code() end

---@return string[]
function Update:args() end

---@return string
function Update:data() end

---@param text? string
---@param alert? boolean
---@return boolean|nil ok
---@return string? err
function Update:answer(text, alert) end

---@return boolean|nil ok
---@return string? err
function Update:delete() end

---@return boolean
function Update:is_reply() end

---@return boolean
function Update:is_bot() end

---@return boolean
function Update:is_outgoing() end

---@return boolean
function Update:is_incoming() end

---@return boolean
function Update:has_message() end

---@return string
function Update:mention() end

---@param chat_id number
---@param text string
---@return boolean|nil ok
---@return string? err
function Update:send_message(chat_id, text) end

---@return boolean
function Update:is_edited() end

---@return string
function Update:connection_id() end

---@return boolean
function Update:is_business() end

---@param method string
---@param params? table
---@return table|nil result
---@return string? err
function Update:raw_call(method, params) end

---@param msg string
---@param ... any
function Update:log_debug(msg, ...) end

---@param msg string
---@param ... any
function Update:log_info(msg, ...) end

---@param msg string
---@param ... any
function Update:log_success(msg, ...) end

---@param msg string
---@param ... any
function Update:log_warn(msg, ...) end

---@param msg string
---@param ... any
function Update:log_error(msg, ...) end

return gotg
