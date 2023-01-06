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

// An example for reading configurations from ENV
func ExampleConfFromEnv() {
	c := conf.New[ExampleConfig](
		conf.WithTagName("json"),   // Tag name must match with the tag name defined inside the struct for unmarshalling the configurations. Default tag name is mapstructure
		conf.WithStores(env.New()), // Create a Store object for reading configurations from ENV
	)

	bc, err := c.Parse()
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(*bc)
}

// An example for reading configurations from local files
func ExampleConfFromFile() {
	c := conf.New[ExampleConfig](
		conf.WithTagName("json"), // Tag name must match with the tag name defined inside the struct for unmarshalling the configurations. Default tag name is mapstructure
		conf.WithStores(file.New(file.WithConfigPaths(file.ConfigPath{Path: "./conf.yaml"}))), // Create a Store object for reading configurations from local files
	)

	bc, err := c.Parse()
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(*bc)
}

// An example for reading configurations from Apollo
func ExampleConfFromApollo() {
	// Create an object for reading Apollo Access Key from a local file.
	// This object can also be used to get configurations to override the configurations from other stores.
	l, err := apollo.NewLocalConfig(
		apollo.WithLocalConfigPath("local_conf.yaml"), // Set path to the local configs. Default is configs.yaml
	)
	if err != nil {
		log.Println(err)
		return
	}

	c := conf.New[ExampleConfig](
		conf.WithTagName("json"),
		conf.WithStores(
			// Create a Store object for reading configurations from Apollo
			apollo.New(
				apollo.WithURL("APOLLO_URL"),     // Set Apollo URL. Default is http://apollo.meta
				apollo.WithAppID("APP_ID"),       // Set App ID
				apollo.WithCluster("CLUSTER_ID"), // Set cluster ID. Default is 'default'
				// Set Apollo Access Key. Default is to read from ENV with key 'APOLLO_ACCESS_KEY' and then 'APOLLO_ACCESSKEY_SECRET'.
				// If not found in ENV, it'll try to read it from LocalConfig (if specified).
				apollo.WithAccessKey("ACCESS_KEY"),
				apollo.WithNamespaces("NS1", "NS2"), // Set namespaces to read configurations from. Default is 'application'.
				apollo.EnableWatch(),                // Enable watching for configuration changes. Default is 'disable'.
				apollo.WithLocalConfig(l),           // Set LocalConfig to read Apollo Access Key from, and override configurations from Apollo. No default LocalConfig.
			),
		),
	)

	bc, err := c.Parse()
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(*bc)

	// Start watching configuration changes.
	// When changed, the latest unmarshalled configuration object and the changes will be returned via the specified callback
	err = c.Watch(func(cfg *ExampleConfig, changes []store.ConfigChange) {
		log.Println(*bc)
		log.Println(changes)
	})

	c.Unwatch() // Stop watching
}

// An example for using template in configurations
func ExampleTemplateData() {
	t, err := tdata.New(
		// Set stores used as data source for replacing templates.
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
		conf.WithTagName("json"),
		conf.WithStores(
			file.New(
				file.WithConfigPaths(file.ConfigPath{Path: "tconf.yaml"}),
				file.WithTemplateData(t), // Use configurations from 't' to replace templates in the above configuration file
			),
		),
	)

	bc, err := c.Parse()
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(*bc)
}
