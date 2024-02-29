namespace WeatherClient
{
    public partial class AppShell : Shell
    {
        public AppShell()
        {
            InitializeComponent();

            Routing.RegisterRoute(nameof(Views.RulesPage), typeof(Views.RulesPage));
        }
    }
}
