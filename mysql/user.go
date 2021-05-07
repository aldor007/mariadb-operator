
package mysql

import (
	"context"
	"errors"
	"fmt"
	mariadbv1alpha1 "github.com/aldor007/mariadb-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"strings"

)

// CreateUserIfNotExists creates a user if it doesn't already exist and it gives it the specified permissions
func CreateUserIfNotExists(ctx context.Context, sql SQLRunner,
	user, pass string, allowedHosts []string, permissions []mariadbv1alpha1.MariaDBPermission,
	resourceLimit mariadbv1alpha1.MariaDBUserLimits
) error {

	// throw error if there are no allowed hosts
	if len(allowedHosts) == 0 {
		return errors.New("no allowedHosts specified")
	}

	queries := []Query{
		getCreateUserQuery(user, pass, allowedHosts),
		getAlterUserQuery(user, pass, allowedHosts, resourceLimit),
	}

	if len(permissions) > 0 {
		queries = append(queries, permissionsToQuery(permissions, user, allowedHosts))
	}

	query := BuildAtomicQuery(queries...)

	if err := sql.QueryExec(ctx, query); err != nil {
		return fmt.Errorf("failed to configure user (user/pass/access), err: %s", err)
	}

	return nil
}

func getAlterUserQuery(user, pwd string, allowedHosts []string, resourceLimit mariadbv1alpha1.MariaDBUserLimits) Query {
	args := []interface{}{}
	q := "ALTER USER"

	// add user identifications (user@allowedHost) pairs
	ids, idsArgs := getUsersIdentification(user, &pwd, allowedHosts)
	q += ids
	args = append(args, idsArgs...)

	// add WITH statement for resource options
	if len(resourceLimit.Get()) > 0 {
		q += " WITH"
		for key, value := range resourceLimit.Get() {
			q += fmt.Sprintf(" %s ?", Escape(string(key)))
			args = append(args, value)
		}
	}

	return NewQuery(q, args...)
}

func getCreateUserQuery(user, pwd string, allowedHosts []string) Query {
	idsTmpl, idsArgs := getUsersIdentification(user, &pwd, allowedHosts)

	return NewQuery(fmt.Sprintf("CREATE USER IF NOT EXISTS%s", idsTmpl), idsArgs...)
}

func getUsersIdentification(user string, pwd *string, allowedHosts []string) (ids string, args []interface{}) {
	for i, host := range allowedHosts {
		// add comma if more than one allowed hosts are used
		if i > 0 {
			ids += ","
		}

		if pwd != nil {
			ids += " ?@? IDENTIFIED BY ?"
			args = append(args, user, host, *pwd)
		} else {
			ids += " ?@?"
			args = append(args, user, host)
		}
	}

	return ids, args
}

// DropUser removes a MySQL user if it exists, along with its privileges
func DropUser(ctx context.Context, sql SQLRunner, user, host string) error {
	query := NewQuery("DROP USER IF EXISTS ?@?;", user, host)

	if err := sql.QueryExec(ctx, query); err != nil {
		return fmt.Errorf("failed to delete user, err: %s", err)
	}

	return nil
}

func permissionsToQuery(permissions []mariadbv1alpha1.MariaDBPermission, user string, allowedHosts []string) Query {
	permQueries := []Query{}

	for _, perm := range permissions {
		// If you wish to grant permissions on all tables, you should explicitly use "*"
		for _, table := range perm.Tables {
			args := []interface{}{}

			escPerms := []string{}
			for _, perm := range perm.Permissions {
				escPerms = append(escPerms, Escape(perm))
			}

			schemaTable := fmt.Sprintf("%s.%s", escapeID(perm.Schema), escapeID(table))

			// Build GRANT query
			idsTmpl, idsArgs := getUsersIdentification(user, nil, allowedHosts)

			query := "GRANT " + strings.Join(escPerms, ", ") + " ON " + schemaTable + " TO" + idsTmpl
			args = append(args, idsArgs...)

			permQueries = append(permQueries, NewQuery(query, args...))
		}
	}

	return ConcatenateQueries(permQueries...)
}
