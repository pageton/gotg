# Spanish FTL locale file for gotg bot
# Formato FTL (Fluent Translation List) con características avanzadas

# Mensajes básicos
welcome = ¡Bienvenido al bot!
goodbye = ¡Adiós! Hasta pronto.
thanks = Gracias por usar este bot.

# Comando de inicio con variables
start =
  ¡Hola { $name }!

  Bienvenido a { $bot }! Soy un bot construido con gotg y gotd.

  Usa /help para ver los comandos disponibles.

# Comando de ayuda
help =
  *Comandos Disponibles:*

  /start - Iniciar el bot
  /help - Mostrar este mensaje de ayuda
  /settings - Cambiar tu configuración
  /language - Cambiar el idioma del bot

# Información del usuario con variables
user-info =
  *Información del Usuario:*
  Nombre: { $name }
  ID de Usuario: { $userID }
  Nombre de Usuario: { $username }

# Ejemplos de pluralización
items-count =
  { $count ->
      [one] Tienes { $count } artículo.
     *[other] Tienes { $count } artículos.
  }

messages-count =
  { $count ->
      [1] Tienes 1 mensaje nuevo.
     *[other] Tienes { $count } mensajes nuevos.
  }

# Mensajes basados en género
greeting =
  { $gender ->
      [male] ¡Hola, guapo!
     [female] ¡Hola, hermosa!
     *[other] ¡Hola!
  }

# Configuración
settings-language = Idioma
settings-notifications = Notificaciones
settings-privacy = Privacidad

# Errores
error-general = Ocurrió un error. Por favor, intenta de nuevo.
error-permission = No tienes permiso para hacer esto.
error-not-found = El recurso solicitado no fue encontrado.

# Mensajes de éxito
success = ¡Éxito!
done = ¡Hecho!
completed = Operación completada exitosamente.

# Botones comunes
btn-yes = Sí
btn-no = No
btn-cancel = Cancelar
btn-back = Atrás
btn-next = Siguiente
btn-menu = Menú

# Selección de idioma
language-select = Por favor, selecciona tu idioma:
language-changed = Idioma cambiado a { $lang }
language-current = Tu idioma actual es: { $lang }

# Características anidadas (usando notación de punto)
features-formatting = Formato de texto con HTML y Markdown
features-i18n = Soporte de internacionalización
features-sessions = Múltiples backends de sesión
features-middleware = Soporte de middleware

# Ejemplo de atributos
share-email =
  .title = Comparte tu email
  .description = Usaremos tu email para enviarte actualizaciones
  .button = Compartir Email

# Mensaje de reserva
key-not-found = Traducción no encontrada: { $key }
