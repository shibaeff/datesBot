package bot

const (
	helpMsg = "🔍 Поиск:\n• /next — просмотреть следующую анкету\n• /Matches — взаимные лайки\n\n📋 Профиль:\n• /profile — посмотреть как выглядит ваш профиль\n• /reregister - пройти регистрацию заново\n• /photo — обновить фото \n• /faculty [имя факультета] — обновить факультет\n• /about [новая информация] — обновить описание \n\n⚙️ Прочие команды:\n• " +
		"/start — общее описание бота\n• /help — вызов этого сообщения\n• /delete — удалить Ваш профиль\n" +
		"• /reset — сбросить все свои оценки (аккуратно!)\n• /feedback [сообщение] — сообщение админам\n\n"
	matchMsg    = "У вас мэтч c %s"
	matchesList = "Вот ваши мэтчи:\n"
	adminHelp   = "Admin commands:\n• /notify [сообщение] - разослать всем сообщение\n• /dump — посмотреть изготовить дамп-файл всей базы" +
		"\n• /users [количество пользователей] — посмотреть пользователей\n /log [n] - залоггировать n последних действий"
)
