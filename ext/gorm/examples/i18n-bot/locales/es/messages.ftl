# Spanish locale file for gotg bot (FTL format)
# Formato FTL con soporte para pluralización, género y contexto

# Mensajes básicos
welcome = ¡Bienvenido al bot!
goodbye = ¡Adiós! Hasta pronto.
thanks = Gracias por usar este bot.

# Comando de inicio
start =
  ¡Hola { $userName }!

  Bienvenido a { $botName }! Soy un bot construido con gotg y gotd.

  Usa /help para ver los comandos disponibles.

# Comando de ayuda
help =
  *Comandos Disponibles:*

  /start - Iniciar el bot
  /help - Mostrar este mensaje de ayuda
  /settings - Cambiar tu configuración
  /language - Cambiar el idioma del bot

# Información del usuario
user_info =
  *Información del Usuario:*
  Nombre: { $name }
  ID de Usuario: { $userId }
  Nombre de Usuario: { $username }

# Ejemplos de pluralización
items-count =
  { $count ->
    [one] Tienes { $count } artículo.
    *[other] Tienes { $count } artículos.
  }

# Configuración
settings_language = Idioma
settings_notifications = Notificaciones
settings_privacy = Privacidad

# Errores
error_general = Ocurrió un error. Por favor, intenta de nuevo.
error_permission = No tienes permiso para hacer esto.
error_not_found = El recurso solicitado no fue encontrado.

# Mensajes de éxito
success = ¡Éxito!
done = ¡Hecho!
completed = Operación completada exitosamente.

# Botones comunes
btn_yes = Sí
btn_no = No
btn_cancel = Cancelar
btn_back = Atrás
btn_next = Siguiente
btn_menu = Menú

# Selección de idioma
language_select = Por favor, selecciona tu idioma:
language_changed = Idioma cambiado a { $lang }
language_current = Tu idioma actual es: { $lang }

# Ejemplo con claves anidadas (usando atributos)
features-formatting = Formato de texto con HTML y Markdown
features-i18n = Soporte de internacionalización
features-sessions = Múltiples backends de sesión
features-middleware = Soporte de middleware
