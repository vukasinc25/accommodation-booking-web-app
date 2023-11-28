import { Component, OnInit } from '@angular/core';
import { Accommodation } from '../model/accommodation';
import { AccommodationService } from '../service/accommodation.service';
import { Router } from '@angular/router';
import { AuthService } from '../service/auth.service';

@Component({
  selector: 'app-main-page',
  templateUrl: './main-page.component.html',
  styleUrls: ['./main-page.component.css'],
})
export class MainPageComponent implements OnInit {
  accommodations: Accommodation[] = [];
  isLoggedin: boolean = false;
  userRole: string = '';

  constructor(
    private router: Router,
    private accommodationService: AccommodationService,
    private authService: AuthService
  ) {
    this.authService.isLoggedin.subscribe((data) => (this.isLoggedin = data));
    this.authService.role.subscribe((data) => (this.userRole = data));
  }

  ngOnInit(): void {
    this.authService.checkLoggin();
    this.authService.checkRole();

    this.accommodationService.getAll().subscribe({
      next: (data) => {
        this.accommodations = data as Accommodation[];
        // console.log(this.accommodations);
      },
      error: (err) => {
        console.log(err);
      },
    });
  }
}
