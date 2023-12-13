import { Component, OnInit } from '@angular/core';
import { AuthService } from '../service/auth.service';

@Component({
  selector: 'app-header',
  templateUrl: './header.component.html',
  styleUrls: ['./header.component.css'],
})
export class HeaderComponent implements OnInit {
  isLoggedin: boolean = false;
  constructor(private authService: AuthService) {
    this.authService.isLoggedin.subscribe((data) => (this.isLoggedin = data));
  }

  ngOnInit(): void {
    this.authService.checkLoggin();
  }
}
