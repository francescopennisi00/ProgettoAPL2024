﻿namespace WeatherClient.Exceptions
{
    internal class EmailAlreadyInUseException : Exception
    {
        public string Errormessage { get; private set; }
        public EmailAlreadyInUseException(string message)
        {
            Errormessage = message;
        }
    }
}
