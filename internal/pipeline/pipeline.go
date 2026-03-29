package pipeline

func Run() {
	// TODO: Инициализация config. Загрузить конфигурацию из YAML или Environment.
	// pathCfg:="config/template/postgres.yaml"
	// cfg, err := config.LoadConfig(pathCfg)
	// if err != nil {
	// 	return err
	// }
	// fmt.Println(cfg)

	// TODO: Анализ дамп-файла. Вызвать DetectBackupType. Если ошибка или тип unknow, то наверное, стоит прекращать работу, нужно подумать для unknow
	/*
		Если заголовок -- PostgreSQL database dump сместится (например, из-за комментариев, добавленных скриптом, или маркеров кодировки UTF-8 BOM),
		100 байт не хватит, и валидный бэкап определится как unknown (Программа читает первые 512 байт). -- Может потом имеет смысл вынести в конфиг
		Так же, если tar файл поврежден он определится как unknown.
	*/
	// dumPath:="Путь до дамп файла"
	// resDet, err:=DetectBackupType(dumPath)
	// if err!=nil{
	// 	return err
	// }

	// TODO: Запуск контейнера с параметрами config.
	// provider, err := StartContainer(ctx, cfg.Docker, cfg.Database)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer provider.Close(ctx)

	// TODO: Установка расширений
	

	// TODO: Если backupType != TypeDumpAll, тоПри необходимости создание и добавление Ролей и Расширений

	// TODO: Восстановление данных при помощи psql or pg_restore

	// TODO: Логическая проверка (Checks из config). Подключение к уже заполненной базе и проверка целостности данных,
	// наличие таблиц, индексов или выполняем кастомные SQL-чеки.

	// TODO: Собрать отчет (log/json) о том, что восстановление прошло успешно.
}
