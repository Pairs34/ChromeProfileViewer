using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Data;
using System.Diagnostics;
using System.Drawing;
using System.IO;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows.Forms;

namespace ChromeProfileViewer
{
    public partial class frmChromeViewer : Form
    {
        public frmChromeViewer()
        {
            InitializeComponent();
        }

        private void lstProfiles_DragEnter(object sender, DragEventArgs e)
        {
            if (e.Data.GetDataPresent(DataFormats.FileDrop)) e.Effect = DragDropEffects.Copy;
        }

        private void lstProfiles_DragDrop(object sender, DragEventArgs e)
        {
            string[] files = (string[])e.Data.GetData(DataFormats.FileDrop);
            foreach (string file in files)
                lstProfiles.Items.Add(file);
        }

        private void btnSelectChrome_Click(object sender, EventArgs e)
        {
            var dialog_result = openFileDialog.ShowDialog();
            if (dialog_result == DialogResult.OK)
            {
                if (String.IsNullOrEmpty(openFileDialog.FileName))
                {
                    txtChromePath.Text = openFileDialog.FileName;
                }
            }
        }

        private void lstProfiles_KeyDown(object sender, KeyEventArgs e)
        {
            if (e.KeyCode == Keys.Delete)
            {
                lstProfiles.Items.RemoveAt(lstProfiles.SelectedIndex);
            }
        }

        private void btnStartProfile_Click(object sender, EventArgs e)
        {
            for (int i = 0; i < lstProfiles.Items.Count; i++)
            {
                ProcessStartInfo pInfo = new ProcessStartInfo();
                pInfo.FileName = txtChromePath.Text;
                pInfo.WorkingDirectory = Path.GetFullPath(txtChromePath.Text);
                pInfo.Arguments = $"--user-data-dir=\"{lstProfiles.Items[i].ToString()}\"";
                Process.Start(pInfo);
            }
        }
    }
}
