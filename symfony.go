package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	OPT_COMPOSER = "--optimize-autoloader --prefer-dist --no-dev --no-interaction --no-progress"
	OPT_COMPOSER_DEV = "--prefer-dist --no-interaction --no-progress"
)

var (
	composer_version,
	symfony_env_prod string
)

func init() {
	symfony_env_prod = "dev"
}

func Symfony(conn *SshConnection, path string) {
	// symfony.assets.update_version
	// symfony.assets.normalize_timestamps
	SymfonyComposerInstall(conn, path)

	// symfony.assets.install
	// symfony.assetic.dump
	// symfony.cache.warmup
	// symfony.project.clear_controllers
	SymfonyDeloySetPermission(conn, path)
}

func SymfonySetup (conn *SshConnection) {
	SymfonyComposerGet(conn, sharedPath)
}

func SymfonyComposerInstall(conn *SshConnection, path string) {
	SymfonyVendorCopy(conn, path)
	if conn.Err != nil {
		return
	}

	defer func() {
		lenText := LogInfo(conn.Host.Host, "Install composer dependencies")
		if conn.Err != nil {
			LogNG(lenText)
			LogError(conn.Host.Host, conn.Err)
			return
		}
		LogOK(lenText)
	}()

	env := make([]string, 0)
	for _, v := range config.Env {
		env = append(env, v)
	}

	composerOpt := OPT_COMPOSER
	if symfony_env_prod == "dev" {
		composerOpt = OPT_COMPOSER_DEV
	}

	conn.Exec(fmt.Sprintf(
		"sh -c 'cd %s && %s SYMFONY_ENV=%s php %s/composer.phar install %s'",
		path,
		strings.Join(env, " "),
		symfony_env_prod,
		sharedPath,
		composerOpt,
	))
}

func SymfonyComposerGet(conn *SshConnection, path string) {
	if conn.Err != nil {
		return
	}

	defer func() {
		lenText := LogInfo(conn.Host.Host, "Gets composer and installs it")
		if conn.Err != nil {
			LogNG(lenText)
			LogError(conn.Host.Host, conn.Err)
			return
		}
		LogOK(lenText)
	}()

	opt := []string{}
	if composer_version != "" {
		opt = append(opt, fmt.Sprintf("-- --version=%s", composer_version))
	}
	if conn.SshIsExist(filepath.Join(path, "composer.phar")) {
		conn.Exec(fmt.Sprintf(
			"sh -c 'cd %s && php composer.phar self-update %s'",
			path,
			composer_version,
		))

	} else {
		conn.Exec(fmt.Sprintf(
			"sh -c 'cd %s && curl -sSL https://getcomposer.org/installer | php %s'",
			path,
			strings.Join(opt, " "),
		))
	}
}

func SymfonyVendorCopy(conn *SshConnection, path string) {
	if conn.Err != nil {
		return
	}

	defer func() {
		lenText := LogInfo(conn.Host.Host, "Copy vendors from previous release")
		if conn.Err != nil {
			LogNG(lenText)
			LogError(conn.Host.Host, conn.Err)
			return
		}
		LogOK(lenText)
	}()

	conn.Exec(fmt.Sprintf(
		"vendorDir=%s/vendor; if [ -d $vendorDir ] || [ -h $vendorDir ]; then cp -a $vendorDir %s; fi;",
		pathCurrent(),
		path,
	))
}

func SymfonyDeloySetPermission(conn *SshConnection, path string) {
	if conn.Err != nil {
		return
	}

	// http://symfony.com/doc/master/book/installation.html#checking-symfony-application-configuration-and-setup
	defer func() {
		lenText := LogInfo(conn.Host.Host, "Sets permissions for writable_dirs")
		if conn.Err != nil {
			LogNG(lenText)
			LogError(conn.Host.Host, conn.Err)
			return
		}
		LogOK(lenText)
	}()

	dirs := []string{}
	for _, v := range writableDirs {
		if StringInSlice(sharedDirs, v) {
			dirs = append(dirs, filepath.Join(sharedPath, v))
		} else {
			dirs = append(dirs, filepath.Join(path, v))
		}
	}

	for _, v := range dirs {
		out := conn.Exec(fmt.Sprintf("stat %s -c %%U", v))
		if conn.Host.User == out {
			out := conn.Exec(fmt.Sprintf(
				"getfacl --absolute-names --tabular %s | grep %s.*rwx | wc -l",
				v,
				webUser,
			))
			if out == "0" {
				conn.Exec(fmt.Sprintf(
					"setfacl -R -m u:%s:rwX -m u:%s:rwX %s",
					conn.Host.User,
					webUser,
					v,
				))
				conn.Exec(fmt.Sprintf(
					"setfacl -dR -m u:%s:rwX -m u:%s:rwX %s",
					conn.Host.User,
					webUser,
					v,
				))
			}
		}
	}
}
