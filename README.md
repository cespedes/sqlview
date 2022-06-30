# sqlview

`sqlview` is a tool to view and modify information from a SQL database
in a text terminal.

It relies on a configuration file, `.sqlview.yaml`,
to specify the SQL queries to use in order to get the results from the database.

When executed, it reads the configuration file, connects to the database and
asks for a query; the results are presented in a text table using a text terminal.
The user can see the results, ask for aditional queries, add,
modify or delete records.

All the behavioud is specified in the configuration file.
