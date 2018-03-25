package models

import (
	"flag"
	"fmt"
	"log"
	"sync"
	//	"time"

	"github.com/go-xorm/xorm"
	_ "github.com/lib/pq"
	"gopkg.in/yaml.v2"
)

var (
	Engine *xorm.Engine
)

func createSchema() error {
	schemas := []string{
		`CREATE EXTENSION "uuid-ossp"
    SCHEMA public
    VERSION "1.1";`, //创建uuid扩展

		`CREATE OR REPLACE FUNCTION public.uuid_generate_v4(
	)
    RETURNS uuid
    LANGUAGE 'c'
    COST 1
    VOLATILE STRICT PARALLEL SAFE
AS '$libdir/uuid-ossp', 'uuid_generate_v4'
;
ALTER FUNCTION public.uuid_generate_v4()
    OWNER TO postgres;`, //创建uuid_generate_v4函数

		`CREATE TABLE public.aliyun_accounts
(
    uuid uuid NOT NULL DEFAULT uuid_generate_v4(),
    ali_uid character varying(16) COLLATE pg_catalog."C",
    created timestamp(0) with time zone
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.aliyun_accounts
    OWNER to postgres;

COMMENT ON COLUMN public.aliyun_accounts.created
    IS '创建时间';

-- Index: aliyun_accounts_indexes

-- DROP INDEX public.aliyun_accounts_indexes;

CREATE INDEX IF NOT EXISTS aliyun_accounts_indexes
    ON public.aliyun_accounts USING brin
    (uuid, ali_uid COLLATE pg_catalog."default")
    TABLESPACE pg_default;`, //创建aliyun_accounts表及表内索引

		`-- Table: public.api_services

-- DROP TABLE public.api_services;

CREATE TABLE public.api_services
(
    account uuid,
 	mobile bigint,
    order_biz_id character varying(16) COLLATE pg_catalog."C",
    order_id character varying(16) COLLATE pg_catalog."C",
    sku_id character varying(16) COLLATE pg_catalog."C",
    has_req bigint DEFAULT 0,
    bill smallint DEFAULT 0,
    created timestamp(0) with time zone,
    expired_on timestamp(0) with time zone,
    is_delete boolean NOT NULL DEFAULT false
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.api_services
    OWNER to postgres;

COMMENT ON COLUMN public.api_services.order_biz_id
    IS '用户购买后生产的业务实例ID';

COMMENT ON COLUMN public.api_services.order_id
    IS '订单ID';

COMMENT ON COLUMN public.api_services.sku_id
    IS '针对商品的某个版本分配的ID';

COMMENT ON COLUMN public.api_services.has_req
    IS '剩余次数';

COMMENT ON COLUMN public.api_services.bill
    IS '计费方式：1，请求次数 2，流量下发';

COMMENT ON COLUMN public.api_services.created
    IS '服务购买时间';

COMMENT ON COLUMN public.api_services.expired_on
    IS '失效日期';`,
	}

	for _, q := range schemas {
		_, err := Engine.Exec(q)
		if err != nil {
			return err
		}
	}
	return nil
}

func init() {
	YamlInit()
	SyncMapMql5 = new(sync.Map)         //mql5 缓存map
	SyncMapSumNetworkIn = new(sync.Map) //统计流量

	var err error
	Engine, err = xorm.NewEngine(Conf.Config.DriverName, Conf.Config.DataSourceName)
	if err != nil {
		fmt.Println(err)
	}

	Engine.ShowExecTime(false)
	Engine.ShowSQL(false)
	Engine.SetMaxOpenConns(10)
	//	tL, err := time.LoadLocation("Asia/Shanghai")
	//	if err != nil {
	//		panic(err)
	//	}
	//	Engine.TZLocation = tL
	err = createSchema()
	if err != nil {
		//panic(err)
	}

}

func YamlInit() {
	flag.Parse()
	b := ReadFile(*yamlFile)
	err := yaml.Unmarshal([]byte(*b), &Conf)
	if err != nil {
		log.Fatalf("readfile(%q): %s", *yamlFile, err)
	}
}
