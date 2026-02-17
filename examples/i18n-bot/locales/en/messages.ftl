# English locale file for gotg bot (FTL format)

welcome = Welcome to the bot!
goodbye = Goodbye! See you soon.
thanks = Thank you for using this bot.

# Start command
start =
  Hello { $userName }!

  Welcome to { $botName }! I'm a bot built with gotg and gotd.

  Use /help to see available commands.

# Help command
help =
  *Available Commands:*

  /start - Start the bot
  /help - Show this help message
  /settings - Change your settings
  /language - Change bot language

# User info
user_info =
  *User Information:*
  Name: { $name }
  User ID: { $userId }
  Username: { $username }

# Pluralization examples
items-count =
  { $count ->
    [one] You have { $count } item.
    *[other] You have { $count } items.
  }

# Settings
settings_language = Language
settings_notifications = Notifications
settings_privacy = Privacy

# Errors
error_general = An error occurred. Please try again.
error_permission = You don't have permission to do this.
error_not_found = The requested resource was not found.

# Success messages
success = Success!
done = Done!
completed = Operation completed successfully.

# Common buttons
btn_yes = Yes
btn_no = No
btn_cancel = Cancel
btn_back = Back
btn_next = Next
btn_menu = Menu

# Language selection
language_select = Please select your language:
language_changed = Language changed to { $lang }
language_current = Your current language is: { $lang }

# Example with nested keys (using attributes)
features-formatting = Text formatting with HTML and Markdown
features-i18n = Internationalization support
features-sessions = Multiple session backends
features-middleware = Middleware support
