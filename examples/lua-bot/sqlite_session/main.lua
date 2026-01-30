local gotg = require("gotg")

local api_id = tonumber(os.getenv("API_ID")) or 0
local api_hash = os.getenv("API_HASH") or ""
local bot_token = os.getenv("BOT_TOKEN") or ""

local client = gotg.new_client({
	api_id = api_id,
	api_hash = api_hash,
	bot_token = bot_token,
	session = "bot.db",
})

client:on_command("start", function(u)
	u:reply("Hello, " .. u:first_name() .. "!")
end, "private")

client:on_command("me", function(u)
	local result, err = u:raw_call("users.getFullUser", {
		id = u:user_id(),
	})
	if err or not result then
		u:reply("Error: " .. (err or "no result"))
		return
	end
	local about = result.full_user and result.full_user.about or "empty"
	u:reply("Bio: " .. about)
end)

client:on_message(function(u)
	u:reply("Echo: " .. u:text())
end, "!command")

client:idle()
