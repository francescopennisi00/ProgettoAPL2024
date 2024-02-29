namespace WeatherClient.Views;

public partial class AllRulesPage : ContentPage
{
    public AllRulesPage()
    {
        InitializeComponent();

    }
    private void ContentPage_NavigatedTo(object sender, NavigatedToEventArgs e)
    {
        notesCollection.SelectedItem = null;
    }
}