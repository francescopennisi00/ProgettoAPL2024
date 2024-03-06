namespace WeatherClient.Exceptions
{
    internal class UsernamePswWrongException : Exception
    {
        public string Errormessage { get; private set; }
        public UsernamePswWrongException(string message)
        {
            Errormessage = message;
        }

    }
}
