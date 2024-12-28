# Check Foreign Key Constraints

This tool validates foreign key constraints in SQL scripts. It ensures that foreign key constraints are only allowed between tables within the same group, as defined in a configuration file.

## Features

- Validates `CREATE TABLE` and `ALTER TABLE` statements for foreign key constraints
- Detects violations where foreign key constraints span tables from different groups
- Outputs detailed logs with file names, line numbers, and suggestions for fixing the issues

## Usage

Run the script with the following command:

```bash
go run ./... -paths ./path/to/dir -config ./path/to/config.json
```

### Parameters

`-paths`: The directory containing the SQL files to analyze. The script recursively checks all .sql files within this directory.  
`-config`: The path to the configuration file defining table groups.

## Configuration File

The configuration file is a JSON file that defines groups of tables. Foreign key constraints are only allowed between tables within the same group.

### Example Configuration

```json
{
  "group1": ["table1", "table2"],
  "group2": ["table3", "table4"],
  "group3": ["table5", "table6"]
}
```

## Example Output

When a foreign key constraint violation is detected, the tool outputs an error log like the following:

```
Foreign key constraint "table1 -> table3" in sql1.sql is not allowed
Foreign key constraint "table1 -> table6" in sql1.sql is not allowed
Foreign key constraint "table3 -> table5" in sql2.sql is not allowed
```
