// Copyright 2020 Author: Ruslan Bikchentaev. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package psql

const (
	sqlTableList = `select TABLE_NAME, TABLE_TYPE
					FROM INFORMATION_SCHEMA.TABLES
					WHERE table_schema = 'public' 
					order by 1`
	sqlFuncList = `select specific_name, routine_name, routine_type, data_type, type_udt_name
					FROM INFORMATION_SCHEMA.routines
					WHERE specific_schema = 'public'`
	sqlGetTablesColumns = `SELECT c.column_name, data_type, COALESCE(column_default, 'NULL'),
								is_nullable='YES', COALESCE(character_set_name, ''),
								COALESCE(character_maximum_length, -1), udt_name,
								k.constraint_name, k.position_in_unique_constraint is null,
								COALESCE(pg_catalog.col_description((SELECT ('"' || $1 || '"')::regclass::oid), c.ordinal_position::int), '')
							   AS column_comment
							FROM INFORMATION_SCHEMA.COLUMNS c
								LEFT JOIN INFORMATION_SCHEMA.key_column_usage k 
									on (k.table_name=c.table_name AND k.column_name = c.COLUMN_NAME)
							WHERE c.table_schema='public' AND c.table_name=$1`
	sqlGetFuncParams = `SELECT parameter_name, data_type, udt_name,
		COALESCE(CHARACTER_SET_NAME, ''),
		COALESCE(CHARACTER_MAXIMUM_LENGTH, -1), COALESCE(parameter_default, 'NULL'),
		ordinal_position, parameter_mode
		FROM INFORMATION_SCHEMA.parameters
		WHERE specific_schema='public' AND specific_name=$1`
	sqlGetColumnAttr = `SELECT data_type, COALESCE(column_default, 'NULL'),
							is_nullable='YES', COALESCE(character_set_name, ''),
							COALESCE(character_maximum_length, -1), udt_name
						FROM INFORMATION_SCHEMA.COLUMNS C
						WHERE C.table_schema='public' AND C.table_name=$1 AND C.COLUMN_NAME = $2`
)
