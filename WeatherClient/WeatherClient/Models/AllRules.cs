using System.Collections.ObjectModel;

namespace WeatherClient.Models;

internal class AllNotes
{
    public ObservableCollection<Rule> Rules { get; set; } = new ObservableCollection<Rule>();

    public AllNotes() =>
        LoadNotes();

    public void LoadNotes()
    {
        Rules.Clear();

        // Get the folder where the notes are stored.
        string appDataPath = FileSystem.AppDataDirectory;

        // Use Linq extensions to load the *.notes.txt files.
        IEnumerable<Rule> notes = Directory

                                    // Select the file names from the directory
                                    .EnumerateFiles(appDataPath, "*.notes.txt")

                                    // Each file name is used to create a new Note
                                    .Select(filename => new Rule()
                                    {
                                        Filename = filename,
                                        Text = File.ReadAllText(filename),
                                        Date = File.GetCreationTime(filename)
                                    })

                                    // With the final collection of notes, order them by date
                                    .OrderBy(note => note.Date);

        // Add each note into the ObservableCollection
        foreach (Rule note in notes)
            Rules.Add(note);
    }
}
