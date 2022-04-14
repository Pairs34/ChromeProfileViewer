namespace ChromeProfileViewer
{
    partial class frmChromeViewer
    {
        /// <summary>
        /// Required designer variable.
        /// </summary>
        private System.ComponentModel.IContainer components = null;

        /// <summary>
        /// Clean up any resources being used.
        /// </summary>
        /// <param name="disposing">true if managed resources should be disposed; otherwise, false.</param>
        protected override void Dispose(bool disposing)
        {
            if (disposing && (components != null))
            {
                components.Dispose();
            }
            base.Dispose(disposing);
        }

        #region Windows Form Designer generated code

        /// <summary>
        /// Required method for Designer support - do not modify
        /// the contents of this method with the code editor.
        /// </summary>
        private void InitializeComponent()
        {
            this.lblChromePath = new System.Windows.Forms.Label();
            this.txtChromePath = new System.Windows.Forms.TextBox();
            this.btnSelectChrome = new System.Windows.Forms.Button();
            this.lstProfiles = new System.Windows.Forms.ListBox();
            this.btnStartProfile = new System.Windows.Forms.Button();
            this.openFileDialog = new System.Windows.Forms.OpenFileDialog();
            this.SuspendLayout();
            // 
            // lblChromePath
            // 
            this.lblChromePath.AutoSize = true;
            this.lblChromePath.Location = new System.Drawing.Point(13, 13);
            this.lblChromePath.Name = "lblChromePath";
            this.lblChromePath.Size = new System.Drawing.Size(68, 13);
            this.lblChromePath.TabIndex = 0;
            this.lblChromePath.Text = "Chrome Path";
            // 
            // txtChromePath
            // 
            this.txtChromePath.Location = new System.Drawing.Point(16, 30);
            this.txtChromePath.Name = "txtChromePath";
            this.txtChromePath.Size = new System.Drawing.Size(683, 20);
            this.txtChromePath.TabIndex = 1;
            this.txtChromePath.Text = "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe";
            // 
            // btnSelectChrome
            // 
            this.btnSelectChrome.Location = new System.Drawing.Point(705, 28);
            this.btnSelectChrome.Name = "btnSelectChrome";
            this.btnSelectChrome.Size = new System.Drawing.Size(75, 23);
            this.btnSelectChrome.TabIndex = 2;
            this.btnSelectChrome.Text = "Seç";
            this.btnSelectChrome.UseVisualStyleBackColor = true;
            this.btnSelectChrome.Click += new System.EventHandler(this.btnSelectChrome_Click);
            // 
            // lstProfiles
            // 
            this.lstProfiles.AllowDrop = true;
            this.lstProfiles.FormattingEnabled = true;
            this.lstProfiles.Location = new System.Drawing.Point(16, 67);
            this.lstProfiles.Name = "lstProfiles";
            this.lstProfiles.Size = new System.Drawing.Size(764, 329);
            this.lstProfiles.TabIndex = 3;
            this.lstProfiles.DragDrop += new System.Windows.Forms.DragEventHandler(this.lstProfiles_DragDrop);
            this.lstProfiles.DragEnter += new System.Windows.Forms.DragEventHandler(this.lstProfiles_DragEnter);
            this.lstProfiles.KeyDown += new System.Windows.Forms.KeyEventHandler(this.lstProfiles_KeyDown);
            // 
            // btnStartProfile
            // 
            this.btnStartProfile.Location = new System.Drawing.Point(705, 404);
            this.btnStartProfile.Name = "btnStartProfile";
            this.btnStartProfile.Size = new System.Drawing.Size(75, 23);
            this.btnStartProfile.TabIndex = 4;
            this.btnStartProfile.Text = "Başlat";
            this.btnStartProfile.UseVisualStyleBackColor = true;
            this.btnStartProfile.Click += new System.EventHandler(this.btnStartProfile_Click);
            // 
            // openFileDialog
            // 
            this.openFileDialog.Title = "Chrome Uygulamasını Seçiniz";
            // 
            // frmChromeViewer
            // 
            this.AutoScaleDimensions = new System.Drawing.SizeF(6F, 13F);
            this.AutoScaleMode = System.Windows.Forms.AutoScaleMode.Font;
            this.ClientSize = new System.Drawing.Size(793, 439);
            this.Controls.Add(this.btnStartProfile);
            this.Controls.Add(this.lstProfiles);
            this.Controls.Add(this.btnSelectChrome);
            this.Controls.Add(this.txtChromePath);
            this.Controls.Add(this.lblChromePath);
            this.FormBorderStyle = System.Windows.Forms.FormBorderStyle.FixedSingle;
            this.MaximizeBox = false;
            this.MinimizeBox = false;
            this.Name = "frmChromeViewer";
            this.Text = "Chrome Profile Viewer";
            this.ResumeLayout(false);
            this.PerformLayout();

        }

        #endregion

        private System.Windows.Forms.Label lblChromePath;
        private System.Windows.Forms.TextBox txtChromePath;
        private System.Windows.Forms.Button btnSelectChrome;
        private System.Windows.Forms.ListBox lstProfiles;
        private System.Windows.Forms.Button btnStartProfile;
        private System.Windows.Forms.OpenFileDialog openFileDialog;
    }
}

