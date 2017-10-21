DROP TABLE IF EXISTS "%[1]d_local_languages"; CREATE TABLE "%[1]d_local_languages" (
  "id" bigint  NOT NULL DEFAULT '0',
  "name" character varying(100) NOT NULL DEFAULT '',
  "res" text NOT NULL DEFAULT ''
);
ALTER TABLE ONLY "%[1]d_local_languages" ADD CONSTRAINT "%[1]d_local_languages_pkey" PRIMARY KEY (id);
CREATE INDEX "%[1]d_local_languages_index_name" ON "%[1]d_local_languages" (name);

DROP TABLE IF EXISTS "%[1]d_local_menu"; CREATE TABLE "%[1]d_local_menu" (
    "id" bigint  NOT NULL DEFAULT '0',
    "name" character varying(255) UNIQUE NOT NULL DEFAULT '',
    "title" character varying(255) NOT NULL DEFAULT '',
    "value" text NOT NULL DEFAULT '',
    "conditions" text NOT NULL DEFAULT ''
);
ALTER TABLE ONLY "%[1]d_local_menu" ADD CONSTRAINT "%[1]d_local_menu_pkey" PRIMARY KEY (id);
CREATE INDEX "%[1]d_local_menu_index_name" ON "%[1]d_local_menu" (name);

DROP TABLE IF EXISTS "%[1]d_local_pages"; CREATE TABLE "%[1]d_local_pages" (
    "id" bigint  NOT NULL DEFAULT '0',
    "name" character varying(255) UNIQUE NOT NULL DEFAULT '',
    "value" text NOT NULL DEFAULT '',
    "menu" character varying(255) NOT NULL DEFAULT '',
    "conditions" text NOT NULL DEFAULT ''
);
ALTER TABLE ONLY "%[1]d_local_pages" ADD CONSTRAINT "%[1]d_local_pages_pkey" PRIMARY KEY (id);
CREATE INDEX "%[1]d_local_pages_index_name" ON "%[1]d_local_pages" (name);

DROP TABLE IF EXISTS "%[1]d_local_blocks"; CREATE TABLE "%[1]d_local_blocks" (
    "id" bigint  NOT NULL DEFAULT '0',
    "name" character varying(255) UNIQUE NOT NULL DEFAULT '',
    "value" text NOT NULL DEFAULT '',
    "conditions" text NOT NULL DEFAULT ''
);
ALTER TABLE ONLY "%[1]d_local_blocks" ADD CONSTRAINT "%[1]d_local_blocks_pkey" PRIMARY KEY (id);
CREATE INDEX "%[1]d_local_blocks_index_name" ON "%[1]d_local_blocks" (name);

DROP TABLE IF EXISTS "%[1]d_local_signatures"; CREATE TABLE "%[1]d_local_signatures" (
    "id" bigint  NOT NULL DEFAULT '0',
    "name" character varying(100) NOT NULL DEFAULT '',
    "value" jsonb,
    "conditions" text NOT NULL DEFAULT ''
);
ALTER TABLE ONLY "%[1]d_local_signatures" ADD CONSTRAINT "%[1]d_local_signatures_pkey" PRIMARY KEY (name);

CREATE TABLE "%[1]d_local_contracts" (
"id" bigint NOT NULL  DEFAULT '0',
"value" text  NOT NULL DEFAULT '',
"conditions" text  NOT NULL DEFAULT ''
);
ALTER TABLE ONLY "%[1]d_local_contracts" ADD CONSTRAINT "%[1]d_local_contracts_pkey" PRIMARY KEY (id);

INSERT INTO "%[1]d_local_contracts" ("id", "value", "wallet_id", "conditions") VALUES 
('1','contract MainCondition {
  conditions {
    if(StateVal("founder_account")!=$wallet)
    {
      warning "Sorry, you don`t have access to this action."
    }
  }
}', '%[2]d', '0', 'ContractConditions(`MainCondition`)');

DROP TABLE IF EXISTS "%[1]d_local_parameters";
CREATE TABLE "%[1]d_local_parameters" (
"id" bigint NOT NULL  DEFAULT '0',
"name" varchar(255) UNIQUE NOT NULL DEFAULT '',
"value" text NOT NULL DEFAULT '',
"conditions" text  NOT NULL DEFAULT ''
);
ALTER TABLE ONLY "%[1]d_local_parameters" ADD CONSTRAINT "%[1]d_local_parameters_pkey" PRIMARY KEY ("id");
CREATE INDEX "%[1]d_local_parameters_index_name" ON "%[1]d_local_parameters" (name);

INSERT INTO "%[1]d_local_parameters" ("id","name", "value", "conditions") VALUES 
('1','founder_account', '%[2]d', 'ContractConditions(`MainCondition`)'),
('2','new_table', 'ContractConditions(`MainCondition`)', 'ContractConditions(`MainCondition`)'),
('3','new_column', 'ContractConditions(`MainCondition`)', 'ContractConditions(`MainCondition`)'),
('4','changing_tables', 'ContractConditions(`MainCondition`)', 'ContractConditions(`MainCondition`)'),
('5','changing_language', 'ContractConditions(`MainCondition`)', 'ContractConditions(`MainCondition`)'),
('6','changing_signature', 'ContractConditions(`MainCondition`)', 'ContractConditions(`MainCondition`)'),
('7','changing_page', 'ContractConditions(`MainCondition`)', 'ContractConditions(`MainCondition`)'),
('8','changing_menu', 'ContractConditions(`MainCondition`)', 'ContractConditions(`MainCondition`)'),
('9','changing_contracts', 'ContractConditions(`MainCondition`)', 'ContractConditions(`MainCondition`)');

CREATE TABLE "%[1]d_local_tables" (
"name" varchar(100) UNIQUE NOT NULL DEFAULT '',
"permissions" jsonb,
"columns" jsonb,
"conditions" text  NOT NULL DEFAULT ''
);
ALTER TABLE ONLY "%[1]d_local_tables" ADD CONSTRAINT "%[1]d_local_tables_pkey" PRIMARY KEY (name);

INSERT INTO "%[1]d_local_tables" ("name", "permissions","columns", "conditions") VALUES ('contracts', 
        '{"insert": "ContractAccess(\"@1NewContract\")", "update": "ContractAccess(\"@1EditContract\")", 
          "new_column": "ContractAccess(\"@1NewColumn\")"}',
        '{"value": "ContractAccess(\"@1EditContract\")",
          "conditions": "ContractAccess(\"@1EditContract\")"}', 'ContractAccess("@1EditTable")'),
        ('languages', 
        '{"insert": "ContractAccess(\"@1NewLang\")", "update": "ContractAccess(\"@1EditLang\")", 
          "new_column": "ContractAccess(\"@1NewColumn\")"}',
        '{ "name": "ContractAccess(\"@1EditLang\")",
          "res": "ContractAccess(\"@1EditLang\")",
          "conditions": "ContractAccess(\"@1EditLang\")"}', 'ContractAccess("@1EditTable")'),
        ('menu', 
        '{"insert": "ContractAccess(\"@1NewMenu\")", "update": "ContractAccess(\"@1EditMenu\")", 
          "new_column": "ContractAccess(\"@1NewColumn\")"}',
        '{"name": "ContractAccess(\"@1EditMenu\")",
    "value": "ContractAccess(\"@1EditMenu\")",
    "conditions": "ContractAccess(\"@1EditMenu\")"
        }', 'ContractAccess("@1EditTable")'),
        ('pages', 
        '{"insert": "ContractAccess(\"@1NewPage\")", "update": "ContractAccess(\"@1EditPage\")", 
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

