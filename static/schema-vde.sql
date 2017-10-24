DROP TABLE IF EXISTS "%[1]d_vde_languages"; CREATE TABLE "%[1]d_vde_languages" (
  "id" bigint  NOT NULL DEFAULT '0',
  "name" character varying(100) NOT NULL DEFAULT '',
  "res" text NOT NULL DEFAULT ''
);
ALTER TABLE ONLY "%[1]d_vde_languages" ADD CONSTRAINT "%[1]d_vde_languages_pkey" PRIMARY KEY (id);
CREATE INDEX "%[1]d_vde_languages_index_name" ON "%[1]d_vde_languages" (name);

DROP TABLE IF EXISTS "%[1]d_vde_menu"; CREATE TABLE "%[1]d_vde_menu" (
    "id" bigint  NOT NULL DEFAULT '0',
    "name" character varying(255) UNIQUE NOT NULL DEFAULT '',
    "title" character varying(255) NOT NULL DEFAULT '',
    "value" text NOT NULL DEFAULT '',
    "conditions" text NOT NULL DEFAULT ''
);
ALTER TABLE ONLY "%[1]d_vde_menu" ADD CONSTRAINT "%[1]d_vde_menu_pkey" PRIMARY KEY (id);
CREATE INDEX "%[1]d_vde_menu_index_name" ON "%[1]d_vde_menu" (name);

DROP TABLE IF EXISTS "%[1]d_vde_pages"; CREATE TABLE "%[1]d_vde_pages" (
    "id" bigint  NOT NULL DEFAULT '0',
    "name" character varying(255) UNIQUE NOT NULL DEFAULT '',
    "value" text NOT NULL DEFAULT '',
    "menu" character varying(255) NOT NULL DEFAULT '',
    "conditions" text NOT NULL DEFAULT ''
);
ALTER TABLE ONLY "%[1]d_vde_pages" ADD CONSTRAINT "%[1]d_vde_pages_pkey" PRIMARY KEY (id);
CREATE INDEX "%[1]d_vde_pages_index_name" ON "%[1]d_vde_pages" (name);

DROP TABLE IF EXISTS "%[1]d_vde_blocks"; CREATE TABLE "%[1]d_vde_blocks" (
    "id" bigint  NOT NULL DEFAULT '0',
    "name" character varying(255) UNIQUE NOT NULL DEFAULT '',
    "value" text NOT NULL DEFAULT '',
    "conditions" text NOT NULL DEFAULT ''
);
ALTER TABLE ONLY "%[1]d_vde_blocks" ADD CONSTRAINT "%[1]d_vde_blocks_pkey" PRIMARY KEY (id);
CREATE INDEX "%[1]d_vde_blocks_index_name" ON "%[1]d_vde_blocks" (name);

DROP TABLE IF EXISTS "%[1]d_vde_signatures"; CREATE TABLE "%[1]d_vde_signatures" (
    "id" bigint  NOT NULL DEFAULT '0',
    "name" character varying(100) NOT NULL DEFAULT '',
    "value" jsonb,
    "conditions" text NOT NULL DEFAULT ''
);
ALTER TABLE ONLY "%[1]d_vde_signatures" ADD CONSTRAINT "%[1]d_vde_signatures_pkey" PRIMARY KEY (name);

CREATE TABLE "%[1]d_vde_contracts" (
"id" bigint NOT NULL  DEFAULT '0',
"value" text  NOT NULL DEFAULT '',
"conditions" text  NOT NULL DEFAULT ''
);
ALTER TABLE ONLY "%[1]d_vde_contracts" ADD CONSTRAINT "%[1]d_vde_contracts_pkey" PRIMARY KEY (id);

DROP TABLE IF EXISTS "%[1]d_vde_parameters";
CREATE TABLE "%[1]d_vde_parameters" (
"id" bigint NOT NULL  DEFAULT '0',
"name" varchar(255) UNIQUE NOT NULL DEFAULT '',
"value" text NOT NULL DEFAULT '',
"conditions" text  NOT NULL DEFAULT ''
);
ALTER TABLE ONLY "%[1]d_vde_parameters" ADD CONSTRAINT "%[1]d_vde_parameters_pkey" PRIMARY KEY ("id");
CREATE INDEX "%[1]d_vde_parameters_index_name" ON "%[1]d_vde_parameters" (name);

INSERT INTO "%[1]d_vde_parameters" ("id","name", "value", "conditions") VALUES 
('1','founder_account', '%[2]d', 'ContractConditions(`MainCondition`)'),
('2','new_table', 'ContractConditions(`MainCondition`)', 'ContractConditions(`MainCondition`)'),
('3','new_column', 'ContractConditions(`MainCondition`)', 'ContractConditions(`MainCondition`)'),
('4','changing_tables', 'ContractConditions(`MainCondition`)', 'ContractConditions(`MainCondition`)'),
('5','changing_language', 'ContractConditions(`MainCondition`)', 'ContractConditions(`MainCondition`)'),
('6','changing_signature', 'ContractConditions(`MainCondition`)', 'ContractConditions(`MainCondition`)'),
('7','changing_page', 'ContractConditions(`MainCondition`)', 'ContractConditions(`MainCondition`)'),
('8','changing_menu', 'ContractConditions(`MainCondition`)', 'ContractConditions(`MainCondition`)'),
('9','changing_contracts', 'ContractConditions(`MainCondition`)', 'ContractConditions(`MainCondition`)');

CREATE TABLE "%[1]d_vde_tables" (
"name" varchar(100) UNIQUE NOT NULL DEFAULT '',
"permissions" jsonb,
"columns" jsonb,
"conditions" text  NOT NULL DEFAULT ''
);
ALTER TABLE ONLY "%[1]d_vde_tables" ADD CONSTRAINT "%[1]d_vde_tables_pkey" PRIMARY KEY (name);

INSERT INTO "%[1]d_vde_tables" ("name", "permissions","columns", "conditions") VALUES ('contracts', 
        '{"insert": "ContractAccess(\"NewContract\")", "update": "ContractAccess(\"EditContract\")", 
          "new_column": "ContractAccess(\"NewColumn\")"}',
        '{"value": "ContractAccess(\"EditContract\")",
          "conditions": "ContractAccess(\"EditContract\")"}', 'ContractAccess("EditTable")'),
        ('languages', 
        '{"insert": "ContractAccess(\"NewLang\")", "update": "ContractAccess(\"EditLang\")", 
          "new_column": "ContractAccess(\"NewColumn\")"}',
        '{ "name": "ContractAccess(\"EditLang\")",
          "res": "ContractAccess(\"EditLang\")",
          "conditions": "ContractAccess(\"EditLang\")"}', 'ContractAccess("EditTable")'),
        ('menu', 
        '{"insert": "ContractAccess(\"NewMenu\")", "update": "ContractAccess(\"EditMenu\")", 
          "new_column": "ContractAccess(\"NewColumn\")"}',
        '{"name": "ContractAccess(\"EditMenu\")",
    "value": "ContractAccess(\"EditMenu\")",
    "conditions": "ContractAccess(\"EditMenu\")"
        }', 'ContractAccess("EditTable")'),
        ('pages', 
        '{"insert": "ContractAccess(\"NewPage\")", "update": "ContractAccess(\"EditPage\")", 
          "new_column": "ContractAccess(\"NewColumn\")"}',
        '{"name": "ContractAccess(\"EditPage\")",
    "value": "ContractAccess(\"EditPage\")",
    "menu": "ContractAccess(\"EditPage\")",
    "conditions": "ContractAccess(\"EditPage\")"
        }', 'ContractAccess("EditTable")'),
        ('blocks', 
        '{"insert": "ContractAccess(\"NewBlock\")", "update": "ContractAccess(\"EditBlock\")", 
          "new_column": "ContractAccess(\"NewColumn\")"}',
        '{"name": "ContractAccess(\"EditBlock\")",
    "value": "ContractAccess(\"EditBlock\")",
    "conditions": "ContractAccess(\"EditBlock\")"
        }', 'ContractAccess("EditTable")'),
        ('signatures', 
        '{"insert": "ContractAccess(\"NewSign\")", "update": "ContractAccess(\"EditSign\")", 
          "new_column": "ContractAccess(\"NewColumn\")"}',
        '{"name": "ContractAccess(\"EditSign\")",
    "value": "ContractAccess(\"EditSign\")",
    "conditions": "ContractAccess(\"EditSign\")"
        }', 'ContractAccess("EditTable")');

INSERT INTO "%[1]d_vde_contracts" ("id", "value", "conditions") VALUES 
('1','contract MainCondition {
  conditions {
    if(EcosystemParam("founder_account")!=$wallet)
    {
      warning "Sorry, you don`t have access to this action."
    }
  }
}', 'ContractConditions(`MainCondition`)'),
('2','contract VDEFunctions {
}

func DBFind(table string).Columns(columns string).Where(where string, params ...)
     .WhereId(id int).Order(order string).Limit(limit int).Offset(offset int).Ecosystem(ecosystem int) array {
    return DBSelect(table, columns, id, order, offset, limit, ecosystem, where, params)
}

func DBString(table, column string, id int) string {
    var ret array
    var result string
    
    ret = DBFind(table).Columns(column).WhereId(id)
    if Len(ret) > 0 {
        var vmap map
        vmap = ret[0]
        result = vmap[column]
    }
    return result
}

func ConditionById(table string, validate bool) {
    var cond string
    cond = DBString(table, `conditions`, $Id)
    if !cond {
        error Sprintf(`Item %%d has not been found`, $Id)
    }
    Eval(cond)
    if validate {
        ValidateCondition($Conditions,$state)
    }
}', 'ContractConditions(`MainCondition`)'),
('3','contract NewContract {
    data {
    	Value      string
    	Conditions string
    }
    conditions {
      ValidateCondition($Conditions,$state)
	    var list array
	    list = ContractsList($Value)
	    var i int
	    while i < Len(list) {
	        if IsContract(list[i], $state) {
	            warning Sprintf(`Contract %%s exists`, list[i] )
	        }
	        i = i + 1
	    }
    }
    action {
        var root, id int
        root = CompileContract($Value, $state, 0, 0)
        id = DBInsert(`contracts`, `value,conditions`, $Value, $Conditions )
        FlushContract(root, id, false)
    }
}', 'ContractConditions(`MainCondition`)'),
('4','contract EditContract {
    data {
        Id         int
    	Value      string
    	Conditions string
    }
    conditions {
        $cur = DBRow(`contracts`, `id,value,conditions`, $Id)
        if Int($cur[`id`]) != $Id {
            error Sprintf(`Contract %%d does not exist`, $Id)
        }
        Eval($cur[`conditions`])
        ValidateCondition($Conditions,$state)
	    var list, curlist array
	    list = ContractsList($Value)
	    curlist = ContractsList($cur[`value`])
	    if Len(list) != Len(curlist) {
	        error `Contracts cannot be removed or inserted`
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
	            error `Contracts names cannot be changed`
	        }
	        i = i + 1
	    }
    }
    action {
        var root int
        root = CompileContract($Value, $state, 0, 0)
        DBUpdate(`contracts`, $Id, `value,conditions`, $Value, $Conditions)
        FlushContract(root, $Id, false)
    }
}', 'ContractConditions(`MainCondition`)');
