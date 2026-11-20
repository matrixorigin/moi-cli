package commands

import (
	"flag"
	"fmt"
	"time"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func addSQLQueryFlags(fs *flag.FlagSet) {
	fs.String("sql-type", "", "SQL query type")
	fs.Int("offset", 0, "Result offset")
	fs.Int("limit", 0, "Result limit")
}

func sqlQueryCLIOptions(fs *flag.FlagSet, databaseFlag string) []sdk.SQLQueryOption {
	var opts []sdk.SQLQueryOption
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case databaseFlag:
			opts = append(opts, sdk.WithSQLQueryDBName(f.Value.String()))
		case "sql-type":
			opts = append(opts, sdk.WithSQLQueryType(f.Value.String()))
		case "offset":
			opts = append(opts, sdk.WithSQLQueryOffset(f.Value.(flag.Getter).Get().(int)))
		case "limit":
			opts = append(opts, sdk.WithSQLQueryLimit(f.Value.(flag.Getter).Get().(int)))
		}
	})
	return opts
}

func addSQLHistoryRangeFlags(fs *flag.FlagSet) {
	fs.String("scope", "", "History category: self or workspace")
	fs.String("start", "", "Start time (RFC3339)")
	fs.String("end", "", "End time (RFC3339)")
}

func addSQLHistoryFilterFlags(fs *flag.FlagSet) {
	fs.Bool("non-user", false, "Include non-user SQL")
	fs.String("sql-source-types", "", "Comma-separated SQL source types")
}

func sqlHistoryTimeRange(fs *flag.FlagSet) (start, end *time.Time, err error) {
	fs.Visit(func(f *flag.Flag) {
		if f.Name != "start" && f.Name != "end" {
			return
		}
		parsed, parseErr := time.Parse(time.RFC3339Nano, f.Value.String())
		if parseErr != nil {
			err = fmt.Errorf("--%s: %w", f.Name, parseErr)
			return
		}
		if f.Name == "start" {
			start = &parsed
		} else {
			end = &parsed
		}
	})
	return
}

func sqlHistoryOverviewCLIOptions(fs *flag.FlagSet) ([]sdk.QueryHistoryOverviewOption, error) {
	start, end, err := sqlHistoryTimeRange(fs)
	if err != nil {
		return nil, err
	}
	var opts []sdk.QueryHistoryOverviewOption
	if start != nil {
		opts = append(opts, sdk.WithQueryHistoryOverviewStart(*start))
	}
	if end != nil {
		opts = append(opts, sdk.WithQueryHistoryOverviewEnd(*end))
	}
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "scope":
			opts = append(opts, sdk.WithQueryHistoryOverviewScope(sdk.QueryHistoryScope(f.Value.String())))
		case "non-user":
			opts = append(opts, sdk.WithQueryHistoryOverviewNonUser(f.Value.(flag.Getter).Get().(bool)))
		case "sql-source-types":
			opts = append(opts, sdk.WithQueryHistoryOverviewSQLSourceTypes(splitCSV(f.Value.String())...))
		}
	})
	return opts, nil
}

func sqlHistoryCLIOptions(fs *flag.FlagSet) ([]sdk.QueryHistoryOption, error) {
	start, end, err := sqlHistoryTimeRange(fs)
	if err != nil {
		return nil, err
	}
	var opts []sdk.QueryHistoryOption
	if start != nil {
		opts = append(opts, sdk.WithQueryHistoryStart(*start))
	}
	if end != nil {
		opts = append(opts, sdk.WithQueryHistoryEnd(*end))
	}
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "scope":
			opts = append(opts, sdk.WithQueryHistoryScope(sdk.QueryHistoryScope(f.Value.String())))
		case "non-user":
			opts = append(opts, sdk.WithQueryHistoryNonUser(f.Value.(flag.Getter).Get().(bool)))
		case "status":
			opts = append(opts, sdk.WithQueryHistoryStatus(f.Value.String()))
		case "users":
			opts = append(opts, sdk.WithQueryHistoryUsers(splitCSV(f.Value.String())...))
		case "databases":
			opts = append(opts, sdk.WithQueryHistoryDatabases(splitCSV(f.Value.String())...))
		case "statement":
			opts = append(opts, sdk.WithQueryHistoryStatement(f.Value.String()))
		case "query-types":
			opts = append(opts, sdk.WithQueryHistoryTypes(splitCSV(f.Value.String())...))
		case "min-duration":
			opts = append(opts, sdk.WithQueryHistoryMinDuration(f.Value.(flag.Getter).Get().(uint64)))
		case "statement-id":
			opts = append(opts, sdk.WithQueryHistoryStatementID(f.Value.String()))
		case "transaction-id":
			opts = append(opts, sdk.WithQueryHistoryTransactionID(f.Value.String()))
		case "session-id":
			opts = append(opts, sdk.WithQueryHistorySessionID(f.Value.String()))
		case "selected-fields":
			opts = append(opts, sdk.WithQueryHistorySelectedFields(splitCSV(f.Value.String())...))
		case "sql-source-types":
			opts = append(opts, sdk.WithQueryHistorySQLSourceTypes(splitCSV(f.Value.String())...))
		case "offset":
			opts = append(opts, sdk.WithQueryHistoryOffset(f.Value.(flag.Getter).Get().(uint)))
		case "limit":
			opts = append(opts, sdk.WithQueryHistoryLimit(f.Value.(flag.Getter).Get().(uint)))
		case "order-by":
			opts = append(opts, sdk.WithQueryHistoryOrderBy(splitCSV(f.Value.String())...))
		case "order":
			opts = append(opts, sdk.WithQueryHistoryOrder(f.Value.String()))
		case "cu":
			opts = append(opts, sdk.WithQueryHistoryCU(f.Value.(flag.Getter).Get().(uint)))
		}
	})
	return opts, nil
}

func sqlRecordCLIOptions(fs *flag.FlagSet) []sdk.QueryRecordOption {
	var opts []sdk.QueryRecordOption
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "scope":
			opts = append(opts, sdk.WithQueryRecordScope(sdk.QueryHistoryScope(f.Value.String())))
		case "start":
			opts = append(opts, sdk.WithQueryRecordStart(f.Value.String()))
		case "end":
			opts = append(opts, sdk.WithQueryRecordEnd(f.Value.String()))
		}
	})
	return opts
}
