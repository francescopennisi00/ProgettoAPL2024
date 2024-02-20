import mysql.connector
from worker.utils.logger import logger


class DatabaseConnector:
    def __init__(self, hostname, port, user, password, database):
        try:
            self._connection = mysql.connector.connect(
                host=hostname,
                port=port,
                user=user,
                password=password,
                database=database
            )
        except mysql.connector.Error as err:
            logger.error("MySQL Exception raised! -> " + str(err) + "\n")
            raise SystemExit

    def execute_query(self, query, params=None, select=True, commit=False):
        try:
            cursor = self._connection.cursor()
            cursor.execute(query, params)
            if not select:
                if commit:
                    cursor.close()
                    if self.commit_update():  # to make changes effective
                        return True
                    else:
                        return False
                else:
                    # in this case insert, delete or update query was executed but commit will be done later
                    cursor.close()
                    return True
            else:
                results = cursor.fetchall()
                cursor.close()
                return results
        except mysql.connector.Error as err:
            logger.error("MySQL Exception raised! -> " + str(err) + "\n")
            return False

    def commit_update(self):
        try:
            self._connection.commit()
            return True
        except mysql.connector.Error as error:
            logger.error("MySQL Exception raised! -> " + str(error) + "\n")
            try:
                self._connection.rollback()
            except Exception as exe:
                logger.error(f" MySQL Exception raised in rollback: {exe}\n")
            return False

    def close(self):
        try:
            self._connection.close()
            return True
        except mysql.connector.Error as error:
            logger.error("MySQL Exception raised! -> " + str(error) + "\n")
        return False
