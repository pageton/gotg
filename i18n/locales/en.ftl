# English FTL locale file for gotg bot
# FTL (Fluent Translation List) format with advanced features

# Basic messages
welcome = Welcome to the bot!
goodbye = Goodbye! See you soon.
thanks = Thank you for using this bot.

# Start command with variables
start =
  Hello { $name }!

  Welcome to { $bot }! I'm a bot built with gotg and gotd.

  Use /help to see available commands.

# Help command
help =
  *Available Commands:*

  /start - Start the bot
  /help - Show this help message
  /settings - Change your settings
  /language - Change bot language

# User info with variables
user-info =
  *User Information:*
  Name: { $name }
  User ID: { $userID }
  Username: { $username }

# Pluralization examples
items-count =
  { $count ->
      [one] You have { $count } item.
     *[other] You have { $count } items.
  }

messages-count =
  { $count ->
      [1] You have 1 new message.
     *[other] You have { $count } new messages.
  }

# Gender-based messages
greeting =
  { $gender ->
      [male] Hello, handsome!
     [female] Hello, beautiful!
     *[other] Hello there!
  }

# Settings
settings-language = Language
settings-notifications = Notifications
settings-privacy = Privacy

# Errors
error-general = An error occurred. Please try again.
error-permission = You don't have permission to do this.
error-not-found = The requested resource was not found.

# Success messages
success = Success!
done = Done!
completed = Operation completed successfully.

# Common buttons
btn-yes = Yes
btn-no = No
btn-cancel = Cancel
btn-back = Back
btn-next = Next
btn-menu = Menu

# Language selection
language-select = Please select your language:
language-changed = Language changed to { $lang }
language-current = Your current language is: { $lang }

# Nested features (using dot notation)
features-formatting = Text formatting with HTML and Markdown
features-i18n = Internationalization support
features-sessions = Multiple session backends
features-middleware = Middleware support

# Attributes example
share-email =
  .title = Share your email
  .description = We'll use your email to send you updates
  .button = Share Email

# Fallback message
key-not-found = Translation not found: { $key }
