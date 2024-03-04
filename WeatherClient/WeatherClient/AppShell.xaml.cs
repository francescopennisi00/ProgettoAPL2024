namespace WeatherClient
{
    public partial class AppShell : Shell
    {
        public AppShell()
        {
            InitializeComponent();

            Routing.RegisterRoute(nameof(Views.RulesPage), typeof(Views.RulesPage));
            Routing.RegisterRoute(nameof(Views.AllRulesPage), typeof(Views.AllRulesPage));
            Routing.RegisterRoute(nameof(Views.LoginPage), typeof(Views.LoginPage));
        }
    }
}
