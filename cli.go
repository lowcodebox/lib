package lib

import (
	"context"
	"os"

	"github.com/urfave/cli"
)

const sep = string(os.PathSeparator)

// RunServiceFuncCLI обраатываем параметры с консоли и вызываем переданую функцию
func RunServiceFuncCLI(ctx context.Context, funcCLI func(ctx context.Context, configfile, dir, port, mode, service, dc, param2, param3, sourcedb, action, version string) error) error {
	var err error

	appCLI := cli.NewApp()
	appCLI.Usage = "Demon Buildbox Proxy started"
	appCLI.Commands = []cli.Command{
		{
			Name: "webinit", ShortName: "",
			Usage: "Start Web-UI from init infractractire LowCodePlatform-service",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "port, c",
					Usage: "Порт запуска UI",
					Value: "8088",
				},
			},
			Action: func(c *cli.Context) error {
				port := c.String("port")

				err = funcCLI(ctx, "", "", port, "", "", "", "", "", "", "webinit", "")
				return err
			},
		},
		{
			Name: "update", ShortName: "",
			Usage: "Update service",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "service, s",
					Usage: "Обновить сервис",
					Value: "lowcodebox",
				},
				cli.StringFlag{
					Name:  "version, v",
					Usage: "Версия, до которой обновляем",
					Value: "latest",
				},
				cli.StringFlag{
					Name:  "arch, a",
					Usage: "386/amd64",
					Value: "",
				},
			},
			Action: func(c *cli.Context) error {
				service := c.String("service")
				version := c.String("version")
				arch := c.String("arch")

				err = funcCLI(ctx, "", "", "", "", service, arch, "", "", "", "update", version)
				return err
			},
		},
		{
			Name: "stop", ShortName: "",
			Usage: "Stop service",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "service, s",
					Usage: "Остановить сервисы (через запятую). '-s systems' - остановить системные сервисы; '-s custom' - остановить рабочие пользовательские сервисы ",
					Value: "all",
				},
			},
			Action: func(c *cli.Context) error {
				service := c.String("service")

				err = funcCLI(ctx, "", "", "", "", service, "", "", "", "", "stop", "")
				return err
			},
		},
		{
			Name: "start", ShortName: "",
			Usage: "Start single Buildbox-service process",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config, c",
					Usage: "Название файла конфигурации, с которым будет запущен сервис",
					Value: "lowcodebox",
				},
				cli.StringFlag{
					Name:  "dir, d",
					Usage: "Путь к шаблонам",
					Value: "",
				},
				cli.StringFlag{
					Name:  "port, p",
					Usage: "Порт, на котором запустить процесс",
					Value: "",
				},
				cli.StringFlag{
					Name:  "mode, m",
					Usage: "Доп.режимы запуска: debug (логирования stdout в файл)",
					Value: "",
				},
				cli.StringFlag{
					Name:  "dc",
					Usage: "Дата-центр, в котором запущен сервис",
					Value: "false",
				},
				cli.StringFlag{
					Name:  "service, s",
					Usage: "Запуск сервиса (для запуска нескольких сервисов укажите их через запятую)",
					Value: "systems",
				},
			},
			Action: func(c *cli.Context) error {
				configfile := c.String("config")
				port := c.String("port")
				dir := c.String("dir")
				dc := c.String("dc")
				mode := c.String("mode")

				service := c.String("service")

				if dir == "default" {
					dir, err = RootDir()
				}

				err = funcCLI(ctx, configfile, dir, port, mode, service, dc, "", "", "", "start", "")
				return err
			},
		},
		{
			Name: "init", ShortName: "",
			Usage: "Init single LowCodePlatform-service process",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "service, s",
					Usage: "Инициализация сервиса",
					Value: "false",
				},
				cli.StringFlag{
					Name:  "version, v",
					Usage: "До какой версии обновить выбранный сервис",
					Value: "latest",
				},
				cli.StringFlag{
					Name:  "dc",
					Usage: "Зарезервировано",
					Value: "false",
				},
				cli.StringFlag{
					Name:  "param2, p2",
					Usage: "Зарезервировано",
					Value: "false",
				},
				cli.StringFlag{
					Name:  "param3, p3",
					Usage: "Зарезервировано",
					Value: "false",
				},
				cli.StringFlag{
					Name:  "dir, d",
					Usage: "Директория создания проекта (по-умолчанию - текущая директория)",
					Value: "",
				},
				cli.StringFlag{
					Name:  "sourcedb, db",
					Usage: "База данных, где будет развернута фабрика (поддерживается SQLite, MySQL, Postgres, CocckroachDB) (по-умолчанию: SQLite)",
					Value: "./default.db",
				},
			},
			Action: func(c *cli.Context) error {
				service := c.String("service")
				dc := c.String("dc")
				param2 := c.String("param2")
				param3 := c.String("param3")
				dir := c.String("dir")
				version := c.String("version")
				sourcedb := c.String("sourcedb")

				if dir == "default" {
					dir, err = RootDir()
				}

				err = funcCLI(ctx, "", dir, "", "", service, dc, param2, param3, sourcedb, "init", version)
				return err
			},
		},
		{
			Name: "secrets", ShortName: "",
			Usage: "Get or set secrets for service",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "action, a",
					Usage: "Действие для секрета. get или set",
					Value: "get",
				},
				cli.StringFlag{
					Name:  "key, k",
					Usage: "Название секрета",
					Value: "nokey",
				},
				cli.StringFlag{
					Name:  "value, v",
					Usage: "Значение секрета для ключа k",
					Value: "novalue",
				},
			},
			Action: func(c *cli.Context) error {
				configfile := c.String("config")
				action := c.String("action")
				key := c.String("key")
				value := c.String("value")

				return funcCLI(ctx, configfile, "", "", "", "", action, key, value, "", "secrets", "")
			},
		},
	}
	err = appCLI.Run(os.Args)

	return err
}

// Stop завершение процесса
func Stop(pid int) (err error) {
	var sig os.Signal
	sig = os.Kill
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	err = p.Signal(sig)
	return err
}

// завершение всех процессов для текущей конфигурации
// config - ид-конфигурации
//func PidsByConfig(config, portProxy string) (result []string, err error) {
//	_, fullresult, _, _ := Ps("full", portProxy)
//
//	// получаем pid для переданной конфигурации
//	for _, v1 := range fullresult {
//		for _, v := range v1 {
//			configfile := v[1] // файл
//			idProcess := v[0]  // pid
//
//			if config == configfile {
//				result = append(result, idProcess)
//			}
//
//			if err != nil {
//				fmt.Println("Error stopped process config:", config, ", err:", err)
//			}
//		}
//	}
//
//	return
//}

// получаем строки пидов подходящих под условия, в котором:
// domain - название проекта (домен)
// alias - название алиас-сервиса (gui/api/proxy и тд - то, что в мап-прокси идет второй частью адреса)
// если алиас явно не задан, то он может быть получен из домена
//func PidsByAlias(domain, alias, portProxy string) (result []string, err error) {
//
//	if domain == "" {
//		domain = "all"
//	}
//	if alias == "" {
//		alias = "all"
//	}
//
//	// можем в домене передать полный путь с учетом алиаса типа buildbox/gui
//	// в этом случае алиас если он явно не задан заполним значением алиаса полученного из домена
//	splitDomain := strings.Split(domain, "/")
//	if len(splitDomain) == 2 {
//		domain = splitDomain[0]
//		alias = splitDomain[1]
//	}
//	_, _, raw, _ := Ps("full", portProxy)
//
//	// получаем pid для переданной конфигурации
//	for _, pidRegistry := range raw {
//		for d, v1 := range pidRegistry {
//			// пропускаем если точное сравнение и не подоходит
//			if domain != "all" && d != domain {
//				continue
//			}
//
//			for a, v2 := range v1 {
//				// пропускаем если точное сравнение и не подоходит
//				if alias != "all" && a != alias {
//					continue
//				}
//
//				for _, v3 := range v2 {
//					k3 := strings.Split(v3, ":")
//					idProcess := k3[0]  // pid
//					// дополняем результат значениями домена и алиаса (для возврата их при остановке если не переданы алиас явно)
//					// бывают значения, когда мы останавлитваем процесс тошько по домену и тогда мы не можем возврашить алиас остановленного процесса
//					// а алиас нужен для поиска в прокси в картах /Pid и /Мар для удаления из активных сервисов по домену и алиасу
//					// если алиаса нет (не приходит в ответе от лоадера, то не находим и прибитые процессы залипают в мапах)
//					result = append(result, v3+":"+ d + ":" + a)
//
//					if err != nil {
//						fmt.Println("Error stopped process: pid:", idProcess, ", err:", err)
//					}
//				}
//			}
//		}
//	}
//
//	return
//}

// уничтожить все процессы
//func Destroy(portProxy string) (err error) {
//	pids, _, _, _ := Ps("pid", portProxy)
//	for _, v := range pids {
//		pi, err := strconv.Atoi(v)
//		if err == nil {
//			Stop(pi)
//		}
//	}
//	return err
//}

// инициализация приложения
//func Install() (err error) {
//
//	// 1. задание переменных окружения
//	currentDir, err := CurrentDir()
//	if err != nil {
//		return
//	}
//	os.Setenv("BBPATH", currentDir)
//
//	//var rootPath = os.Getenv("BBPATH")
//
//	//fmt.Println(rootPath)
//	//path, _ := os.LookupEnv("BBPATH")
//	//fmt.Print("BBPATH: ", path)
//
//	// 2. копирование файла запуска в /etc/bin
//	//src := "./buildbox"
//	//dst := "/usr/bin/buildbox"
//	//
//	//in, err := os.Open(src)
//	//if err != nil {
//	//	return err
//	//}
//	//defer in.Close()
//	//
//	//out, err := os.Create(dst)
//	//if err != nil {
//	//	return err
//	//}
//	//defer out.Close()
//	//
//	//_, err = io.Copy(out, in)
//	//if err != nil {
//	//	return err
//	//}
//	//return out.Close()
//
//	return err
//}
