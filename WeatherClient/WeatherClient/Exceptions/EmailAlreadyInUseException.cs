using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace WeatherClient.Exceptions
{
    internal class EmailAlreadyInUseException : Exception
    {
        public string Errormessage { get; private set; }
        public EmailAlreadyInUseException(string message)
        {
            this.Errormessage = message;
        }
    }
}
