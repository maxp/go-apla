package migration

import (
	"time"

	"github.com/AplaProject/go-apla/packages/model"
	version "github.com/hashicorp/go-version"
)

var (
	SchemaEcosystem = `DROP TABLE IF EXISTS "%[1]d_keys"; CREATE TABLE "%[1]d_keys" (
	"id" bigint  NOT NULL DEFAULT '0',
	"pub" bytea  NOT NULL DEFAULT '',
	"amount" decimal(30) NOT NULL DEFAULT '0',
	"rb_id" bigint NOT NULL DEFAULT '0'
	);
	ALTER TABLE ONLY "%[1]d_keys" ADD CONSTRAINT "%[1]d_keys_pkey" PRIMARY KEY (id);
	
	DROP TABLE IF EXISTS "%[1]d_history"; CREATE TABLE "%[1]d_history" (
	"id" bigint NOT NULL  DEFAULT '0',
	"sender_id" bigint NOT NULL DEFAULT '0',
	"recipient_id" bigint NOT NULL DEFAULT '0',
	"amount" decimal(30) NOT NULL DEFAULT '0',
	"comment" text NOT NULL DEFAULT '',
	"block_id" int  NOT NULL DEFAULT '0',
	"txhash" bytea  NOT NULL DEFAULT '',
	"rb_id" int  NOT NULL DEFAULT '0'
	);
	ALTER TABLE ONLY "%[1]d_history" ADD CONSTRAINT "%[1]d_history_pkey" PRIMARY KEY (id);
	CREATE INDEX "%[1]d_history_index_sender" ON "%[1]d_history" (sender_id);
	CREATE INDEX "%[1]d_history_index_recipient" ON "%[1]d_history" (recipient_id);
	CREATE INDEX "%[1]d_history_index_block" ON "%[1]d_history" (block_id, txhash);
	
	
	DROP TABLE IF EXISTS "%[1]d_languages"; CREATE TABLE "%[1]d_languages" (
	  "id" bigint  NOT NULL DEFAULT '0',
	  "name" character varying(100) NOT NULL DEFAULT '',
	  "res" jsonb,
	  "conditions" text NOT NULL DEFAULT '',
	  "rb_id" bigint NOT NULL DEFAULT '0'
	);
	ALTER TABLE ONLY "%[1]d_languages" ADD CONSTRAINT "%[1]d_languages_pkey" PRIMARY KEY (id);
	CREATE INDEX "%[1]d_languages_index_name" ON "%[1]d_languages" (name);
	
	DROP TABLE IF EXISTS "%[1]d_menu"; CREATE TABLE "%[1]d_menu" (
		"id" bigint  NOT NULL DEFAULT '0',
		"name" character varying(255) UNIQUE NOT NULL DEFAULT '',
		"value" text NOT NULL DEFAULT '',
		"conditions" text NOT NULL DEFAULT '',
		"rb_id" bigint NOT NULL DEFAULT '0'
	);
	ALTER TABLE ONLY "%[1]d_menu" ADD CONSTRAINT "%[1]d_menu_pkey" PRIMARY KEY (id);
	CREATE INDEX "%[1]d_menu_index_name" ON "%[1]d_menu" (name);
	
	DROP TABLE IF EXISTS "%[1]d_pages"; CREATE TABLE "%[1]d_pages" (
		"id" bigint  NOT NULL DEFAULT '0',
		"name" character varying(255) UNIQUE NOT NULL DEFAULT '',
		"value" text NOT NULL DEFAULT '',
		"menu" character varying(255) NOT NULL DEFAULT '',
		"conditions" text NOT NULL DEFAULT '',
		"rb_id" bigint NOT NULL DEFAULT '0'
	);
	ALTER TABLE ONLY "%[1]d_pages" ADD CONSTRAINT "%[1]d_pages_pkey" PRIMARY KEY (id);
	CREATE INDEX "%[1]d_pages_index_name" ON "%[1]d_pages" (name);
	
	DROP TABLE IF EXISTS "%[1]d_blocks"; CREATE TABLE "%[1]d_blocks" (
		"id" bigint  NOT NULL DEFAULT '0',
		"name" character varying(255) UNIQUE NOT NULL DEFAULT '',
		"value" text NOT NULL DEFAULT '',
		"conditions" text NOT NULL DEFAULT '',
		"rb_id" bigint NOT NULL DEFAULT '0'
	);
	ALTER TABLE ONLY "%[1]d_blocks" ADD CONSTRAINT "%[1]d_blocks_pkey" PRIMARY KEY (id);
	CREATE INDEX "%[1]d_blocks_index_name" ON "%[1]d_blocks" (name);
	
	DROP TABLE IF EXISTS "%[1]d_signatures"; CREATE TABLE "%[1]d_signatures" (
		"id" bigint  NOT NULL DEFAULT '0',
		"name" character varying(100) NOT NULL DEFAULT '',
		"value" jsonb,
		"conditions" text NOT NULL DEFAULT '',
		"rb_id" bigint NOT NULL DEFAULT '0'
	);
	ALTER TABLE ONLY "%[1]d_signatures" ADD CONSTRAINT "%[1]d_signatures_pkey" PRIMARY KEY (name);
	
	CREATE TABLE "%[1]d_contracts" (
	"id" bigint NOT NULL  DEFAULT '0',
	"value" text  NOT NULL DEFAULT '',
	"wallet_id" bigint NOT NULL DEFAULT '0',
	"token_id" bigint NOT NULL DEFAULT '1',
	"active" character(1) NOT NULL DEFAULT '0',
	"conditions" text  NOT NULL DEFAULT '',
	"rb_id" bigint NOT NULL DEFAULT '0'
	);
	ALTER TABLE ONLY "%[1]d_contracts" ADD CONSTRAINT "%[1]d_contracts_pkey" PRIMARY KEY (id);
	
	INSERT INTO "%[1]d_contracts" ("id", "value", "wallet_id","active", "conditions") VALUES 
	('1','contract MainCondition {
	  conditions {
		if(StateVal("founder_account")!=$citizen)
		{
		  warning "Sorry, you don't have access to this action."
		}
	  }
	}', '%[2]d', '0', 'ContractConditions("MainCondition")');
	
	DROP TABLE IF EXISTS "%[1]d_parameters";
	CREATE TABLE "%[1]d_parameters" (
	"id" bigint NOT NULL  DEFAULT '0',
	"name" varchar(255) NOT NULL DEFAULT '',
	"value" text NOT NULL DEFAULT '',
	"conditions" text  NOT NULL DEFAULT '',
	"rb_id" bigint  NOT NULL DEFAULT '0'
	);
	ALTER TABLE ONLY "%[1]d_parameters" ADD CONSTRAINT "%[1]d_parameters_pkey" PRIMARY KEY ("id");
	CREATE INDEX "%[1]d_parameters_index_name" ON "%[1]d_parameters" (name);
	
	INSERT INTO "%[1]d_parameters" ("id","name", "value", "conditions") VALUES 
	('1','founder_account', '%[2]d', 'ContractConditions("MainCondition")'),
	('2','full_node_wallet_id', '%[2]d', 'ContractConditions("MainCondition")'),
	('3','host', '', 'ContractConditions("MainCondition")'),
	('4','restore_access_condition', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
	('5','new_table', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
	('6','new_column', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
	('7','changing_tables', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
	('8','changing_language', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
	('9','changing_signature', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
	('10','changing_page', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
	('11','changing_menu', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
	('12','changing_contracts', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
	('13','ecosystem_name', '%[3]s', 'ContractConditions("MainCondition")'),
	('14','max_sum', '1000000', 'ContractConditions("MainCondition")'),
	('15','citizenship_cost', '1', 'ContractConditions("MainCondition")'),
	('16','money_digit', '2', 'ContractConditions("MainCondition")');
	
	CREATE TABLE "%[1]d_tables" (
	"name" varchar(100) UNIQUE NOT NULL DEFAULT '',
	"permissions" jsonb,
	"columns" jsonb,
	"conditions" text  NOT NULL DEFAULT '',
	"rb_id" bigint NOT NULL DEFAULT '0'
	);
	ALTER TABLE ONLY "%[1]d_tables" ADD CONSTRAINT "%[1]d_tables_pkey" PRIMARY KEY (name);
	
	INSERT INTO "%[1]d_tables" ("name", "permissions","columns", "conditions") VALUES ('contracts', 
			'{"insert": "ContractAccess(\"@1NewContract\")", "update": "ContractAccess(\"@1EditContract\")", 
			  "new_column": "ContractAccess(\"@1NewColumn\")"}',
			'{"value": "ContractAccess(\"@1EditContract\", \"@1ActivateContract\")",
			  "wallet_id": "ContractAccess(\"@1EditContract\", \"@1ActivateContract\")",
			  "token_id": "ContractAccess(\"@1EditContract\", \"@1ActivateContract\")",
			  "active": "ContractAccess(\"@1EditContract\", \"@1ActivateContract\")",
			  "conditions": "ContractAccess(\"@1EditContract\", \"@1ActivateContract\")"}', 'ContractAccess("@1EditTable")'),
			('keys', 
			'{"insert": "ContractAccess(\"@1MoneyTransfer\", \"@1NewEcosystem\")", "update": "ContractAccess(\"@1MoneyTransfer\")", 
			  "new_column": "ContractAccess(\"@1NewColumn\")"}',
			'{"pub": "ContractAccess(\"@1MoneyTransfer\")",
			  "amount": "ContractAccess(\"@1MoneyTransfer\")"}', 'ContractAccess("@1EditTable")'),
			('history', 
			'{"insert": "ContractAccess(\"@1MoneyTransfer\")", "update": "false", 
			  "new_column": "false"}',
			'{"sender_id": "ContractAccess(\"@1MoneyTransfer\")",
			  "recipient_id": "ContractAccess(\"@1MoneyTransfer\")",
			  "amount":  "ContractAccess(\"@1MoneyTransfer\")",
			  "comment": "ContractAccess(\"@1MoneyTransfer\")",
			  "block_id":  "ContractAccess(\"@1MoneyTransfer\")",
			  "txhash": "ContractAccess(\"@1MoneyTransfer\")"}', 'ContractAccess("@1EditTable")'),        
			('languages', 
			'{"insert": "ContractAccess(\"@1NewLang\")", "update": "ContractAccess(\"@1EditLang\")", 
			  "new_column": "ContractAccess(\"@1NewColumn\")"}',
			'{ "name": "ContractAccess(\"@1EditLang\")",
			  "res": "ContractAccess(\"@1EditLang\")",
			  "conditions": "ContractAccess(\"@1EditLang\")"}', 'ContractAccess("@1EditTable")'),
			('menu', 
			'{"insert": "ContractAccess(\"@1NewMenu\", \"@1NewEcosystem\")", "update": "ContractAccess(\"@1EditMenu\")", 
			  "new_column": "ContractAccess(\"@1NewColumn\")"}',
			'{"name": "ContractAccess(\"@1EditMenu\")",
		"value": "ContractAccess(\"@1EditMenu\")",
		"conditions": "ContractAccess(\"@1EditMenu\")"
			}', 'ContractAccess("@1EditTable")'),
			('pages', 
			'{"insert": "ContractAccess(\"@1NewPage\", \"@1NewEcosystem\")", "update": "ContractAccess(\"@1EditPage\")", 
			  "new_column": "ContractAccess(\"@1NewColumn\")"}',
			'{"name": "ContractAccess(\"@1EditPage\")",
		"value": "ContractAccess(\"@1EditPage\")",
		"menu": "ContractAccess(\"@1EditPage\")",
		"conditions": "ContractAccess(\"@1EditPage\")"
			}', 'ContractAccess("@1EditTable")'),
			('blocks', 
			'{"insert": "ContractAccess(\"@1NewBlock\")", "update": "ContractAccess(\"@1EditBlock\")", 
			  "new_column": "ContractAccess(\"@1NewColumn\")"}',
			'{"name": "ContractAccess(\"@1EditBlock\")",
		"value": "ContractAccess(\"@1EditBlock\")",
		"conditions": "ContractAccess(\"@1EditBlock\")"
			}', 'ContractAccess("@1EditTable")'),
			('signatures', 
			'{"insert": "ContractAccess(\"@1NewSign\")", "update": "ContractAccess(\"@1EditSign\")", 
			  "new_column": "ContractAccess(\"@1NewColumn\")"}',
			'{"name": "ContractAccess(\"@1EditSign\")",
		"value": "ContractAccess(\"@1EditSign\")",
		"conditions": "ContractAccess(\"@1EditSign\")"
			}', 'ContractAccess("@1EditTable")');
	
	`
	SchemaFirstEcosystem = `INSERT INTO "system_states" ("id","rb_id") VALUES ('1','0');
	
	INSERT INTO "1_contracts" ("id","value", "wallet_id", "conditions") VALUES 
	('2','contract MoneyTransfer {
		data {
			Recipient string
			Amount    string
			Comment     string "optional"
		}
		conditions {
			$recipient = AddressToId($Recipient)
			if $recipient == 0 {
				error Sprintf("Recipient %%s is invalid", $Recipient)
			}
			var total money
			$amount = Money($Amount) 
			if $amount == 0 {
				error "Amount is zero"
			}
			total = Money(DBString(Table("keys"), "amount", $wallet))
			if $amount >= total {
				error Sprintf("Money is not enough %%v < %%v",total, $amount)
			}
		}
		action {
			DBUpdate(Table("keys"), $wallet,"-amount", $amount)
			DBUpdate(Table("keys"), $recipient,"+amount", $amount)
			DBInsert(Table("history"), "sender_id,recipient_id,amount,comment,block_id,txhash", 
				$wallet, $recipient, $amount, $Comment, $block, $txhash)
		}
	}', '%[1]d', 'ContractConditions("MainCondition")'),
	('3','contract NewContract {
		data {
			Value      string
			Conditions string
			Wallet         string "optional"
			TokenEcosystem int "optional"
		}
		conditions {
			ValidateCondition($Conditions,$state)
			$walletContract = $wallet
			   if $Wallet {
				$walletContract = AddressToId($Wallet)
				if $walletContract == 0 {
				   error Sprintf("wrong wallet %%s", $Wallet)
				}
			}
			var list array
			list = ContractsList($Value)
			var i int
			while i < Len(list) {
				if IsContract(list[i], $state) {
					warning Sprintf("Contract %%s exists", list[i] )
				}
				i = i + 1
			}
			if !$TokenEcosystem {
				$TokenEcosystem = 1
			} else {
				if !SysFuel($TokenEcosystem) {
					warning Sprintf("Ecosystem %%d is not system", $TokenEcosystem )
				}
			}
		}
		action {
			var root, id int
			root = CompileContract($Value, $state, $walletContract, $TokenEcosystem)
			id = DBInsert(Table("contracts"), "value,conditions, wallet_id, token_id", 
				   $Value, $Conditions, $walletContract, $TokenEcosystem)
			FlushContract(root, id, false)
		}
		func price() int {
			return  SysParamInt("contract_price")
		}
	}', '%[1]d', 'ContractConditions("MainCondition")'),
	('4','contract EditContract {
		data {
			Id         int
			Value      string
			Conditions string
		}
		conditions {
			$cur = DBRow(Table("contracts"), "id,value,conditions,active,wallet_id,token_id", $Id)
			if Int($cur["id"]) != $Id {
				error Sprintf("Contract %%d does not exist", $Id)
			}
			Eval($cur["conditions"])
			ValidateCondition($Conditions,$state)
			var list, curlist array
			list = ContractsList($Value)
			curlist = ContractsList($cur["value"])
			if Len(list) != Len(curlist) {
				error "Contracts cannot be removed or inserted"
			}
			var i int
			while i < Len(list) {
				var j int
				var ok bool
				while j < Len(curlist) {
					if curlist[j] == list[i] {
						ok = true
						break
					}
					j = j + 1 
				}
				if !ok {
					error "Contracts names cannot be changed"
				}
				i = i + 1
			}
		}
		action {
			var root int
			root = CompileContract($Value, $state, Int($cur["wallet_id"]), Int($cur["token_id"]))
			DBUpdate(Table("contracts"), $Id, "value,conditions", $Value, $Conditions)
			FlushContract(root, $Id, Int($cur["active"]) == 1)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('5','contract ActivateContract {
		data {
			Id         int
		}
		conditions {
			$cur = DBRow(Table("contracts"), "id,conditions,active,wallet_id", $Id)
			if Int($cur["id"]) != $Id {
				error Sprintf("Contract %%d does not exist", $Id)
			}
			if Int($cur["active"]) == 1 {
				error Sprintf("The contract %%d has been already activated", $Id)
			}
			Eval($cur["conditions"])
			if $wallet != Int($cur["wallet_id"]) {
				error Sprintf("Wallet %%d cannot activate the contract", $wallet)
			}
		}
		action {
			DBUpdate(Table("contracts"), $Id, "active", 1)
			Activate($Id, $state)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('6','contract NewEcosystem {
		data {
			Name  string "optional"
		}
		conditions {
			if $Name && FindEcosystem($Name) {
				error Sprintf("Ecosystem %%s is already existed", $Name)
			}
		}
		action {
			var id int
			id = CreateEcosystem($wallet, $Name)
			DBInsert(Str(id) + "_pages", "name,value,menu,conditions", "default_page", 
				  SysParamString("default_ecosystem_page"), "default_menu", "ContractConditions("MainCondition")")
			DBInsert(Str(id) + "_menu", "name,value,conditions", "default_menu", 
				  SysParamString("default_ecosystem_menu"), "ContractConditions("MainCondition")")
			DBInsert(Str(id) + "_keys", "id,pub", $wallet, DBString("1_keys", "pub", $wallet))
			$result = id
		}
		func price() int {
			return  SysParamInt("ecosystem_price")
		}
		func rollback() {
			RollbackEcosystem()
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('7','contract NewParameter {
		data {
			Name string
			Value string
			Conditions string
		}
		conditions {
			ValidateCondition($Conditions, $state)
		}
		action {
			DBInsert(Table("parameters"), "name,value,conditions", $Name, $Value, $Conditions )
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('8','contract EditParameter {
		data {
			Name string
			Value string
			Conditions string
		}
		conditions {
			EvalCondition(Table("parameters"), $Name, "conditions")
			ValidateCondition($Conditions, $state)
			var exist int
			   if $Name == "ecosystem_name" {
				exist = FindEcosystem($Value)
				if exist > 0 && exist != $state {
					warning Sprintf("Ecosystem %%s already exists", $Value)
				}
			}
		}
		action {
			DBUpdateExt(Table("parameters"), "name", $Name, "value,conditions", $Value, $Conditions )
			   if $Name == "ecosystem_name" {
				DBUpdate("system_states", $state, "name", $Value)
			}
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('9', 'contract NewMenu {
		data {
			Name       string
			Value      string
			Conditions string
		}
		conditions {
			ValidateCondition($Conditions,$state)
		}
		action {
			DBInsert(Table("menu"), "name,value,conditions", $Name, $Value, $Conditions )
		}
		func price() int {
			return  SysParamInt("menu_price")
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('10','contract EditMenu {
		data {
			Id         int
			Value      string
			Conditions string
		}
		conditions {
			Eval(DBString(Table("menu"), "conditions", $Id))
			ValidateCondition($Conditions,$state)
		}
		action {
			DBUpdate(Table("menu"), $Id, "value,conditions", $Value, $Conditions)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('11','contract AppendMenu {
		data {
			Id     int
			Value      string
		}
		conditions {
			Eval(DBString(Table("menu"), "conditions", $Id ))
		}
		action {
			var table string
			table = Table("menu")
			DBUpdate(table, $Id, "value", DBString(table, "value", $Id) + "\r\n" + $Value )
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('12','contract NewPage {
		data {
			Name       string
			Value      string
			Menu       string
			Conditions string
		}
		conditions {
			ValidateCondition($Conditions,$state)
			   if HasPrefix($Name, "sys-") || HasPrefix($Name, "app-") {
				error "The name cannot start with sys- or app-"
			}
		}
		action {
			DBInsert(Table("pages"), "name,value,menu,conditions", $Name, $Value, $Menu, $Conditions )
		}
		func price() int {
			return  SysParamInt("page_price")
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('13','contract EditPage {
		data {
			Id         int
			Value      string
			Menu      string
			Conditions string
		}
		conditions {
			Eval(DBString(Table("pages"), "conditions", $Id))
			ValidateCondition($Conditions,$state)
		}
		action {
			DBUpdate(Table("pages"), $Id, "value,menu,conditions", $Value, $Menu, $Conditions)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('14','contract AppendPage {
		data {
			Id         int
			Value      string
		}
		conditions {
			Eval(DBString(Table("pages"), "conditions", $Id))
		}
		action {
			var value, table string
			table = Table("pages")
			value = DBString(table, "value", $Id)
			   if Contains(value, "PageEnd:") {
			   value = Replace(value, "PageEnd:", $Value) + "\r\nPageEnd:"
			} else {
				value = value + "\r\n" + $Value
			}
			DBUpdate(table, $Id, "value",  value )
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('15','contract NewLang {
		data {
			Name  string
			Trans string
		}
		conditions {
			EvalCondition(Table("parameters"), "changing_language", "value")
			var exist string
			exist = DBStringExt(Table("languages"), "name", $Name, "name")
			if exist {
				error Sprintf("The language resource %%s already exists", $Name)
			}
		}
		action {
			DBInsert(Table("languages"), "name,res", $Name, $Trans )
			UpdateLang($Name, $Trans)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('16','contract EditLang {
		data {
			Name  string
			Trans string
		}
		conditions {
			EvalCondition(Table("parameters"), "changing_language", "value")
		}
		action {
			DBUpdateExt(Table("languages"), "name", $Name, "res", $Trans )
			UpdateLang($Name, $Trans)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('17','contract NewSign {
		data {
			Name       string
			Value      string
			Conditions string
		}
		conditions {
			ValidateCondition($Conditions,$state)
			var exist string
			exist = DBStringExt(Table("signatures"), "name", $Name, "name")
			if exist {
				error Sprintf("The signature %%s already exists", $Name)
			}
		}
		action {
			DBInsert(Table("signatures"), "name,value,conditions", $Name, $Value, $Conditions )
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('18','contract EditSign {
		data {
			Id         int
			Value      string
			Conditions string
		}
		conditions {
			Eval(DBString(Table("signatures"), "conditions", $Id))
			ValidateCondition($Conditions,$state)
		}
		action {
			DBUpdate(Table("signatures"), $Id, "value,conditions", $Value, $Conditions)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('19','contract RequestCitizenship {
		data {
			Name      string
		}
		conditions {
		}
		action {
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('20','contract NewBlock {
		data {
			Name       string
			Value      string
			Conditions string
		}
		conditions {
			ValidateCondition($Conditions,$state)
			   if HasPrefix($Name, "sys-") || HasPrefix($Name, "app-") {
				error "The name cannot start with sys- or app-"
			}
		}
		action {
			DBInsert(Table("blocks"), "name,value,conditions", $Name, $Value, $Conditions )
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('21','contract EditBlock {
		data {
			Id         int
			Value      string
			Conditions string
		}
		conditions {
			Eval(DBString(Table("blocks"), "conditions", $Id))
			ValidateCondition($Conditions,$state)
		}
		action {
			DBUpdate(Table("blocks"), $Id, "value,conditions", $Value, $Conditions)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('22','contract NewTable {
		data {
			Name       string
			Columns      string
			Permissions string
		}
		conditions {
			TableConditions($Name, $Columns, $Permissions)
		}
		action {
			CreateTable($Name, $Columns, $Permissions)
		}
		func rollback() {
			RollbackTable($Name)
		}
		func price() int {
			return  SysParamInt("table_price")
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('23','contract EditTable {
		data {
			Name       string
			Permissions string
		}
		conditions {
			TableConditions($Name, "", $Permissions)
		}
		action {
			PermTable($Name, $Permissions )
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('24','contract NewColumn {
		data {
			TableName   string
			Name        string
			Type        string
			Permissions string
			Index       string "optional"
		}
		conditions {
			ColumnCondition($TableName, $Name, $Type, $Permissions, $Index)
		}
		action {
			CreateColumn($TableName, $Name, $Type, $Permissions, $Index)
		}
		func rollback() {
			RollbackColumn($TableName, $Name)
		}
		func price() int {
			return  SysParamInt("column_price")
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('25','contract EditColumn {
		data {
			TableName   string
			Name        string
			Permissions string
		}
		conditions {
			ColumnCondition($TableName, $Name, "", $Permissions, "")
		}
		action {
			PermColumn($TableName, $Name, $Permissions)
		}
	}', '%[1]d','ContractConditions("MainCondition")');`

	Schema = `DROP TABLE IF EXISTS "transactions_status"; CREATE TABLE "transactions_status" (
		"hash" bytea  NOT NULL DEFAULT '',
		"time" int NOT NULL DEFAULT '0',
		"type" int NOT NULL DEFAULT '0',
		"ecosystem" int NOT NULL DEFAULT '1',
		"wallet_id" bigint NOT NULL DEFAULT '0',
		"block_id" int NOT NULL DEFAULT '0',
		"error" varchar(255) NOT NULL DEFAULT ''
		);
		ALTER TABLE ONLY "transactions_status" ADD CONSTRAINT transactions_status_pkey PRIMARY KEY (hash);
		
		DROP TABLE IF EXISTS "confirmations"; CREATE TABLE "confirmations" (
		"block_id" bigint  NOT NULL DEFAULT '0',
		"good" int  NOT NULL DEFAULT '0',
		"bad" int  NOT NULL DEFAULT '0',
		"time" int  NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "confirmations" ADD CONSTRAINT confirmations_pkey PRIMARY KEY (block_id);
		
		DROP TABLE IF EXISTS "block_chain"; CREATE TABLE "block_chain" (
		"id" int NOT NULL DEFAULT '0',
		"hash" bytea  NOT NULL DEFAULT '',
		"data" bytea NOT NULL DEFAULT '',
		"state_id" int  NOT NULL DEFAULT '0',
		"wallet_id" bigint  NOT NULL DEFAULT '0',
		"time" int NOT NULL DEFAULT '0',
		"tx" int NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "block_chain" ADD CONSTRAINT block_chain_pkey PRIMARY KEY (id);
		
		DROP TABLE IF EXISTS "log_transactions"; CREATE TABLE "log_transactions" (
		"hash" bytea  NOT NULL DEFAULT '',
		"time" int NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "log_transactions" ADD CONSTRAINT log_transactions_pkey PRIMARY KEY (hash);
		
		DROP TABLE IF EXISTS "main_lock"; CREATE TABLE "main_lock" (
		"lock_time" int  NOT NULL DEFAULT '0',
		"script_name" varchar(100) NOT NULL DEFAULT '',
		"info" text NOT NULL DEFAULT '',
		"uniq" smallint NOT NULL DEFAULT '0'
		);
		CREATE UNIQUE INDEX main_lock_uniq ON "main_lock" USING btree (uniq);
		
		DROP TABLE IF EXISTS "migration_history"; CREATE TABLE "migration_history" (
		"id" int NOT NULL  DEFAULT '0',
		"version" int NOT NULL DEFAULT '0',
		"date_applied" int NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "migration_history" ADD CONSTRAINT migration_history_pkey PRIMARY KEY (id);
		
		DROP TABLE IF EXISTS "queue_tx"; CREATE TABLE "queue_tx" (
		"hash" bytea  NOT NULL DEFAULT '',
		"data" bytea NOT NULL DEFAULT '',
		"from_gate" int NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "queue_tx" ADD CONSTRAINT queue_tx_pkey PRIMARY KEY (hash);
		
		DROP TABLE IF EXISTS "config"; CREATE TABLE "config" (
		"my_block_id" int NOT NULL DEFAULT '0',
		"dlt_wallet_id" bigint NOT NULL DEFAULT '0',
		"state_id" int NOT NULL DEFAULT '0',
		"citizen_id" bigint NOT NULL DEFAULT '0',
		"bad_blocks" text NOT NULL DEFAULT '',
		"auto_reload" int NOT NULL DEFAULT '0',
		"first_load_blockchain_url" varchar(255)  NOT NULL DEFAULT '',
		"first_load_blockchain"  varchar(255)  NOT NULL DEFAULT '',
		"current_load_blockchain"  varchar(255)  NOT NULL DEFAULT ''
		);
		
		DROP SEQUENCE IF EXISTS rollback_rb_id_seq CASCADE;
		CREATE SEQUENCE rollback_rb_id_seq START WITH 1;
		DROP TABLE IF EXISTS "rollback"; CREATE TABLE "rollback" (
		"rb_id" bigint NOT NULL  default nextval('rollback_rb_id_seq'),
		"block_id" bigint NOT NULL DEFAULT '0',
		"data" text NOT NULL DEFAULT ''
		);
		ALTER SEQUENCE rollback_rb_id_seq owned by rollback.rb_id;
		ALTER TABLE ONLY "rollback" ADD CONSTRAINT rollback_pkey PRIMARY KEY (rb_id);
		
		DROP TABLE IF EXISTS "system_states"; CREATE TABLE "system_states" (
		"id" bigint NOT NULL DEFAULT '0',
		"name" varchar(255) NOT NULL DEFAULT '',
		"rb_id" bigint NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "system_states" ADD CONSTRAINT system_states_pkey PRIMARY KEY (id);
		CREATE INDEX "system_states_index_name" ON "system_states" (name);
		
		DROP TABLE IF EXISTS "system_parameters";
		CREATE TABLE "system_parameters" (
		"name" varchar(255)  NOT NULL DEFAULT '',
		"value" text NOT NULL DEFAULT '',
		"conditions" text  NOT NULL DEFAULT '',
		"rb_id" bigint  NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "system_parameters" ADD CONSTRAINT system_parameters_pkey PRIMARY KEY ("name");
		
		INSERT INTO system_parameters ("name", "value", "conditions") VALUES 
		('default_ecosystem_page', 'P(class, Default Ecosystem Page)', 'ContractAccess("@0UpdSysParam")'),    
		('default_ecosystem_menu', 'MenuItem(main, Default Ecosystem Menu)', 'ContractAccess("@0UpdSysParam")'),
		('default_ecosystem_contract', '', 'ContractAccess("@0UpdSysParam")'),
		('gap_between_blocks', '3', 'ContractAccess("@0UpdSysParam")'),
		('new_version_url', 'upd.apla.io', 'ContractAccess("@0UpdSysParam")'),
		('full_nodes', '', 'ContractAccess("@0UpdFullNodes")'),
		('count_of_nodes', '101', 'ContractAccess("@0UpdSysParam")'),
		('op_price', '', 'ContractAccess("@0UpdSysParam")'),
		('ecosystem_price', '1000', 'ContractAccess("@0UpdSysParam")'),
		('contract_price', '200', 'ContractAccess("@0UpdSysParam")'),
		('column_price', '200', 'ContractAccess("@0UpdSysParam")'),
		('table_price', '200', 'ContractAccess("@0UpdSysParam")'),
		('menu_price', '100', 'ContractAccess("@0UpdSysParam")'),
		('page_price', '100', 'ContractAccess("@0UpdSysParam")'),
		('blockchain_url', '', 'ContractAccess("@0UpdSysParam")'),
		('max_block_size', '67108864', 'ContractAccess("@0UpdSysParam")'),
		('max_tx_size', '33554432', 'ContractAccess("@0UpdSysParam")'),
		('max_tx_count', '1000', 'ContractAccess("@0UpdSysParam")'),
		('max_columns', '50', 'ContractAccess("@0UpdSysParam")'),
		('max_indexes', '1', 'ContractAccess("@0UpdSysParam")'),
		('max_block_user_tx', '100', 'ContractAccess("@0UpdSysParam")'),
		('max_fuel_tx', '1000', 'ContractAccess("@0UpdSysParam")'),
		('max_fuel_block', '100000', 'ContractAccess("@0UpdSysParam")'),
		('upd_full_nodes_period', '3600', 'ContractAccess("@0UpdSysParam")'),
		('last_upd_full_nodes', '23672372', 'ContractAccess("@0UpdSysParam")'),
		('size_price', '100', 'ContractAccess("@0UpdSysParam")'),
		('commission_size', '3', 'ContractAccess("@0UpdSysParam")'),
		('commission_wallet', '[["1","8275283526439353759"]]', 'ContractAccess("@0UpdSysParam")'),
		('sys_currencies', '[1]', 'ContractAccess("@0UpdSysParam")'),
		('fuel_rate', '[["1","1000000000000000"]]', 'ContractAccess("@0UpdSysParam")'),
		('recovery_address', '[["1","8275283526439353759"]]', 'ContractAccess("@0UpdSysParam")');
		
		CREATE TABLE "system_contracts" (
		"id" bigint NOT NULL  DEFAULT '0',
		"value" text  NOT NULL DEFAULT '',
		"wallet_id" bigint NOT NULL DEFAULT '0',
		"token_id" bigint NOT NULL DEFAULT '0',
		"active" character(1) NOT NULL DEFAULT '0',
		"conditions" text  NOT NULL DEFAULT '',
		"rb_id" bigint NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "system_contracts" ADD CONSTRAINT system_contracts_pkey PRIMARY KEY (id);
		
		INSERT INTO system_contracts ("id","value", "active", "conditions") VALUES 
		('1','contract UpdSysParam {
			data {
			}
			conditions {
			}
			action {
			}
		}', '0','ContractAccess("@0UpdSysContract")'),
		('2','contract UpdSysContract {
			data {
			}
			conditions {
			}
			action {
			}
		}', '0','ContractAccess("@0UpdSysContract")'),
		('3','contract UpdFullNodes {
			 data {
			}
			conditions {
			  var prev int
			  var nodekey bytes
			  prev = DBInt("upd_full_nodes", "time", 1)
				if $time-prev < SysParamInt("upd_full_nodes_period") {
					warning Sprintf("txTime - upd_full_nodes < UPD_FULL_NODES_PERIOD")
				}
			}
			action {
			}
		}', '0','ContractAccess("@0UpdSysContract")');
		
		CREATE TABLE "upd_contracts" (
		"id" bigint NOT NULL  DEFAULT '0',
		"id_contract" bigint  NOT NULL DEFAULT '0',
		"value" text  NOT NULL DEFAULT '',
		"votes" bigint  NOT NULL DEFAULT '0',
		"rb_id" bigint NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "upd_contracts" ADD CONSTRAINT upd_contracts_pkey PRIMARY KEY (id);
		
		CREATE TABLE "upd_system_parameters" (
		"id" bigint NOT NULL DEFAULT '0',
		"name" varchar(255)  NOT NULL DEFAULT '',
		"value" text  NOT NULL DEFAULT '',
		"votes" bigint  NOT NULL DEFAULT '0',
		"rb_id" bigint NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "upd_system_parameters" ADD CONSTRAINT upd_system_parameters_pkey PRIMARY KEY (id);
		
		CREATE TABLE "system_tables" (
		"name" varchar(100)  NOT NULL DEFAULT '',
		"permissions" jsonb,
		"columns" jsonb,
		"conditions" text  NOT NULL DEFAULT '',
		"rb_id" bigint NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "system_tables" ADD CONSTRAINT system_tables_pkey PRIMARY KEY (name);
		
		INSERT INTO system_tables ("name", "permissions","columns", "conditions") VALUES ('upd_contracts', 
				'{"insert": "ContractAccess(\"@0UpdSysContract\")", "update": "ContractAccess(\"@0UpdSysContract\")", 
				  "new_column": "ContractAccess(\"@0UpdSysContract\")"}',
				'{"id_contract": "ContractAccess(\"@0UpdSysContract\")", "value": "ContractAccess(\"@0UpdSysContract\")", 
				  "votes": "ContractAccess(\"@0UpdSysContract\")"}',          
				'ContractAccess(\"@0UpdSysContract\")'),
				('upd_system_parameters', 
				'{"insert": "ContractAccess(\"@0UpdSysContract\")", "update": "ContractAccess(\"@0UpdSysContract\")", 
				  "new_column": "ContractAccess(\"@0UpdSysContract\")"}',
				'{"name": "ContractAccess(\"@0UpdSysContract\")", "value": "ContractAccess(\"@0UpdSysContract\")", 
				  "votes": "ContractAccess(\"@0UpdSysContract\")"}',          
				'ContractAccess(\"@0UpdSysContract\")'),
				('system_states', 
				'{"insert": "false", "update": "ContractAccess(\"@1EditParameter\")", 
				  "new_column": "false"}',
				'{"name": "ContractAccess(\"@1EditParameter\")"}',          
				'ContractAccess(\"@0UpdSysContract\")');
		
		
		DROP TABLE IF EXISTS "info_block"; CREATE TABLE "info_block" (
		"hash" bytea  NOT NULL DEFAULT '',
		"block_id" int NOT NULL DEFAULT '0',
		"state_id" int  NOT NULL DEFAULT '0',
		"wallet_id" bigint NOT NULL DEFAULT '0',
		"time" int  NOT NULL DEFAULT '0',
		"level" smallint  NOT NULL DEFAULT '0',
		"current_version" varchar(50) NOT NULL DEFAULT '0.0.1',
		"sent" smallint NOT NULL DEFAULT '0'
		);
		
		DROP TABLE IF EXISTS "queue_blocks"; CREATE TABLE "queue_blocks" (
		"hash" bytea  NOT NULL DEFAULT '',
		"full_node_id" int NOT NULL DEFAULT '0',
		"block_id" int NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "queue_blocks" ADD CONSTRAINT queue_blocks_pkey PRIMARY KEY (hash);
		
		DROP TABLE IF EXISTS "transactions"; CREATE TABLE "transactions" (
		"hash" bytea  NOT NULL DEFAULT '',
		"data" bytea NOT NULL DEFAULT '',
		"used" smallint NOT NULL DEFAULT '0',
		"high_rate" smallint NOT NULL DEFAULT '0',
		"type" smallint NOT NULL DEFAULT '0',
		"ecosystem" int NOT NULL DEFAULT '1',
		"wallet_id" bigint NOT NULL DEFAULT '0',
		"citizen_id" bigint NOT NULL DEFAULT '0',
		"counter" smallint NOT NULL DEFAULT '0',
		"sent" smallint NOT NULL DEFAULT '0',
		"verified" smallint NOT NULL DEFAULT '1'
		);
		ALTER TABLE ONLY "transactions" ADD CONSTRAINT transactions_pkey PRIMARY KEY (hash);
		
		DROP SEQUENCE IF EXISTS rollback_tx_id_seq CASCADE;
		CREATE SEQUENCE rollback_tx_id_seq START WITH 1;
		DROP TABLE IF EXISTS "rollback_tx"; CREATE TABLE "rollback_tx" (
		"id" bigint NOT NULL  default nextval('rollback_tx_id_seq'),
		"block_id" bigint NOT NULL DEFAULT '0',
		"tx_hash" bytea  NOT NULL DEFAULT '',
		"table_name" varchar(255) NOT NULL DEFAULT '',
		"table_id" varchar(255) NOT NULL DEFAULT ''
		);
		ALTER SEQUENCE rollback_tx_id_seq owned by rollback_tx.id;
		ALTER TABLE ONLY "rollback_tx" ADD CONSTRAINT rollback_tx_pkey PRIMARY KEY (id);
		
		DROP TABLE IF EXISTS "install"; CREATE TABLE "install" (
		"progress" varchar(10) NOT NULL DEFAULT ''
		);
		
		
		DROP TYPE IF EXISTS "my_node_keys_enum_status" CASCADE;
		CREATE TYPE "my_node_keys_enum_status" AS ENUM ('my_pending','approved');
		DROP SEQUENCE IF EXISTS my_node_keys_id_seq CASCADE;
		CREATE SEQUENCE my_node_keys_id_seq START WITH 1;
		DROP TABLE IF EXISTS "my_node_keys"; CREATE TABLE "my_node_keys" (
		"id" int NOT NULL  default nextval('my_node_keys_id_seq'),
		"add_time" int NOT NULL DEFAULT '0',
		"public_key" bytea  NOT NULL DEFAULT '',
		"private_key" varchar(3096) NOT NULL DEFAULT '',
		"status" my_node_keys_enum_status  NOT NULL DEFAULT 'my_pending',
		"my_time" int NOT NULL DEFAULT '0',
		"time" bigint NOT NULL DEFAULT '0',
		"block_id" int NOT NULL DEFAULT '0',
		"rb_id" int NOT NULL DEFAULT '0'
		);
		ALTER SEQUENCE my_node_keys_id_seq owned by my_node_keys.id;
		ALTER TABLE ONLY "my_node_keys" ADD CONSTRAINT my_node_keys_pkey PRIMARY KEY (id);
		
		DROP TABLE IF EXISTS "stop_daemons"; CREATE TABLE "stop_daemons" (
		"stop_time" int NOT NULL DEFAULT '0'
		);
		
		DROP SEQUENCE IF EXISTS full_nodes_id_seq CASCADE;
		CREATE SEQUENCE full_nodes_id_seq START WITH 1;
		DROP TABLE IF EXISTS "full_nodes"; CREATE TABLE "full_nodes" (
		"id" int NOT NULL  default nextval('full_nodes_id_seq'),
		"host" varchar(100) NOT NULL DEFAULT '',
		"wallet_id" bigint  NOT NULL DEFAULT '0',
		"state_id" int NOT NULL DEFAULT '0',
		"final_delegate_wallet_id" bigint NOT NULL DEFAULT '0',
		"final_delegate_state_id" bigint NOT NULL DEFAULT '0',
		"rb_id" int NOT NULL DEFAULT '0'
		);
		ALTER SEQUENCE full_nodes_id_seq owned by full_nodes.id;
		ALTER TABLE ONLY "full_nodes" ADD CONSTRAINT full_nodes_pkey PRIMARY KEY (id);
		
		DROP SEQUENCE IF EXISTS upd_full_nodes_id_seq CASCADE;
		CREATE SEQUENCE upd_full_nodes_id_seq START WITH 1;
		DROP TABLE IF EXISTS "upd_full_nodes"; CREATE TABLE "upd_full_nodes" (
		"id" bigint NOT NULL  default nextval('upd_full_nodes_id_seq'),
		"time" int NOT NULL DEFAULT '0',
		"rb_id" bigint  REFERENCES rollback(rb_id) NOT NULL DEFAULT '0'
		);
		ALTER SEQUENCE upd_full_nodes_id_seq owned by upd_full_nodes.id;
		ALTER TABLE ONLY "upd_full_nodes" ADD CONSTRAINT upd_full_nodes_pkey PRIMARY KEY (id);

		DROP TABLE IF EXISTS "migration"; CREATE TABLE "migration" (
			"version" varchar(100)  NOT NULL DEFAULT ''
			);
		`
)

type migrationData struct {
	vers      *version.Version
	migration string
}

var versionedMigrations []migrationData

func init() {
	versionedMigrations = make([]migrationData, 0)
	/*
		version1, _ := version.NewVersion("1.0.1")
		migration1 := migrationData{version1, `CREATE TABLE "migration_test" (
			"name" varchar(100)  NOT NULL DEFAULT '',
			"permissions" jsonb,
			"columns" jsonb,
			"conditions" text  NOT NULL DEFAULT '',
			"rb_id" bigint NOT NULL DEFAULT '0'
			);`}

		version2, _ := version.NewVersion("2.0.1")
		migration2 := migrationData{version2, `CREATE TABLE "migrations_test_2" (
			"name" varchar(100)  NOT NULL DEFAULT '',
			"permissions" jsonb,
			"columns" jsonb,
			"conditions" text  NOT NULL DEFAULT '',
			"rb_id" bigint NOT NULL DEFAULT '0'
			);`}
	*/
	versionedMigrations = append(versionedMigrations, migration1)
	versionedMigrations = append(versionedMigrations, migration2)
}

func Migrate(vers *version.Version) err {
	for _, migration := range versionedMigrations {
		if migration.Vers.LessThan(vers) {
			err := model.DBConn.Exec(migration.migration)
			if err != nil {
				return err
			}
			dbMigration := &model.MigrationHistory{Version: migration.vers.String(), DateApplied: time.Date().Now()}
			err = dbMigration.Save()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
