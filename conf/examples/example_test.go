/*
 *
 * Copyright (C) 2023 Antigloss Huang (https://github.com/antigloss) All rights reserved.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package examples_test

import (
	"log"

	"github.com/antigloss/go/conf"
	"github.com/antigloss/go/conf/store"
	"github.com/antigloss/go/conf/store/apollo"
	"github.com/antigloss/go/conf/store/env"
	"github.com/antigloss/go/conf/store/file"
	"github.com/antigloss/go/conf/tdata"
)

type ExampleConfig struct {
	Path   string
	Locale struct {
		DefaultLanguage string `json:"default_language"`
	}
	ApolloConfig struct {
		Endpoint   string `json:"end_point"`
		AppID      string `json:"app_id"`
		Cluster    string
		AppSecret  string `json:"app_secret"`
		Namespaces []string
	} `json:"apollo_config"`
}

// 从 Env 读取配置的示例
func ExampleConfFromEnv() {
	c := conf.New[ExampleConfig](
		conf.WithTagName("json"),   // 设置反序列化时使用的 TagName ，默认是 mapstructure
		conf.WithStores(env.New()), // 创建从 Env 读取配置的 Store 对象，并传入 ConfigParser
	)

	bc, err := c.Parse()
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(*bc)
}

// 从文件中读取配置的示例
func ExampleConfFromFile() {
	c := conf.New[ExampleConfig](
		conf.WithTagName("json"), // 设置反序列化时使用的 TagName ，默认是 mapstructure
		conf.WithStores(file.New(file.WithConfigPaths(file.ConfigPath{Path: "./conf.yaml"}))), // 创建从文件中读取配置的 Store 对象，并传入 ConfigParser
	)

	bc, err := c.Parse()
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(*bc)
}

// 从 Apollo 中读取配置的示例
func ExampleConfFromApollo() {
	// 创建本地配置对象，store.apollo 可以使用该对象获取 Apollo 访问秘钥，并且会使用其中的配置项，覆盖 Apollo 上的同名配置
	l, err := apollo.NewLocalConfig(
		apollo.WithLocalConfigPath("local_conf.yaml"), // 设置本地覆盖文件路径，不传默认为 configs.yaml
	)
	if err != nil {
		log.Println(err)
		return
	}

	c := conf.New[ExampleConfig](
		conf.WithTagName("json"), // 设置反序列化时使用的 TagName ，默认是 mapstructure
		conf.WithStores(
			// 创建从 Apollo 读取配置的 Store 对象，并传入 ConfigParser
			apollo.New(
				apollo.WithURL("APOLLO_URL"),     // 设置 Apollo 地址，不传默认为 http://apollo.meta
				apollo.WithAppID("APP_ID"),       // 设置 app id
				apollo.WithCluster("CLUSTER_ID"), // 设置使用的 cluster ，不传默认为 default
				// 设置 Apollo 访问秘钥，不传的话，会尝试从环境变量 APOLLO_ACCESS_KEY 和 APOLLO_ACCESSKEY_SECRET 中获取。
				// 如果环境变量中没有，并且设置了 LocalConfig ，则还会尝试从 LocalConfig 中获取。
				apollo.WithAccessKey("ACCESS_KEY"),
				apollo.WithNamespaces("NS1", "NS2"), // 设置从哪些 Namespaces 中读取配置，不传默认为 application
				apollo.EnableWatch(),                // 监听 Apollo 配置变化，默认是不监听
				apollo.WithLocalConfig(l),           // 设置本地覆盖文件，不传则不使用本地覆盖
			),
		),
	)

	bc, err := c.Parse()
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(*bc)

	// 开始监听 Apollo 配置变化。配置变化后，会调用回调函数，返回最新的配置对象 cfg ，并通过 changes 告知哪些字段发生了变化
	err = c.Watch(func(cfg *ExampleConfig, changes []store.ConfigChange) {
		log.Println(*bc)
		log.Println(changes)
	})

	c.Unwatch() // 取消监听
}

// 使用 TemplateData 替换配置中的模板（text/template）参数
func ExampleTemplateData() {
	t, err := tdata.New(
		// 设置 TemplateData 使用的 Store 对象。支持 env、file 和 apollo ，支持同时设置多个
		tdata.WithStores(
			env.New(),
			apollo.New(),
		),
	)
	if err != nil {
		log.Println(err)
		return
	}

	c := conf.New[ExampleConfig](
		conf.WithTagName("json"), // 设置反序列化时使用的 TagName ，默认是 mapstructure
		conf.WithStores(
			file.New(
				file.WithConfigPaths(file.ConfigPath{Path: "tconf.yaml"}),
				file.WithTemplateData(t), // 使用 t 的数据，替换配置中的模板参数
			),
		),
	)

	bc, err := c.Parse() // 解析配置时，会自动使用 t 的数据，替换配置值的模板参数
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(*bc)

	// 开始监听 Apollo 配置变化。配置变化后，会调用回调函数，返回最新的配置对象 cfg ，并通过 changes 告知哪些字段发生了变化
	err = c.Watch(func(cfg *ExampleConfig, changes []store.ConfigChange) {
		log.Println(*bc)
		log.Println(changes)
	})

	c.Unwatch() // 取消监听
}
