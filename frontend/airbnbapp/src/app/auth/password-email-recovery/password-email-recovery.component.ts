import { Component } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { AuthService } from '../../service/auth.service';
import { Router } from '@angular/router';

@Component({
  selector: 'app-password-email-recovery',
  templateUrl: './password-email-recovery.component.html',
  styleUrls: ['./password-email-recovery.component.css'],
})
export class PasswordEmailRecoveryComponent {
  emailForm: FormGroup;
  constructor(
    private fb: FormBuilder,
    private router: Router,
    private authService: AuthService
  ) {
    this.emailForm = this.fb.group({
      email: ['', [Validators.required, Validators.email]],
    });
  }
  sendVerificationEmail() {
    if (this.emailForm.valid) {
      const email = this.emailForm.get('email')?.value;
      this.authService.sendForgottenPasswordEmail(email).subscribe({
        next: (data) => {
          console.log('Email sent successfully');
          alert('Email sent successfully.');
          this.router.navigate(['reset-password']);
        },
        error: (err) => {
          console.log(err.error.message);
          alert(err.error.message);
        },
      });
    } else {
      console.log('Form is invalid');
    }
  }
}
