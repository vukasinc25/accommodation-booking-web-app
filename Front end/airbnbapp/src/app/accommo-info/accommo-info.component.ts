import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { AccommodationService } from '../service/accommodation.service';
import { Accommodation } from '../model/accommodation';
import { NgbDate } from '@ng-bootstrap/ng-bootstrap';
import { AuthService } from '../service/auth.service';

@Component({
  selector: 'app-accommo-info',
  templateUrl: './accommo-info.component.html',
  styleUrls: ['./accommo-info.component.css'],
})
export class AccommoInfoComponent implements OnInit {
  constructor(
    private route: ActivatedRoute,
    private accommodationService: AccommodationService,
    private authService: AuthService
  ) {}

  role: string = '';
  username: string = '';
  id: number = 0;
  accommodation: Accommodation = {};
  isDataEmpty = false;

  hoveredDate: NgbDate | null = null;
  fromDate: NgbDate | null = null;
  toDate: NgbDate | null = null;

  ngOnInit(): void {
    this.route.params.subscribe((params) => {
      this.id = params['id'];
    });
    this.role = this.authService.getRole();
    this.username = this.authService.getUsername();

    this.accommodationService.getById(this.id).subscribe({
      next: (data) => {
        this.accommodation = data;
        // console.log(data);
      },
      error: (err) => {
        console.log(err);
        this.isDataEmpty = true;
      },
    });
  }

  onDateSelection(date: NgbDate) {
    if (!this.fromDate && !this.toDate) {
      this.fromDate = date;
    } else if (this.fromDate && !this.toDate && date.after(this.fromDate)) {
      this.toDate = date;
    } else {
      this.toDate = null;
      this.fromDate = date;
    }
  }

  isHovered(date: NgbDate) {
    return (
      this.fromDate &&
      !this.toDate &&
      this.hoveredDate &&
      date.after(this.fromDate) &&
      date.before(this.hoveredDate)
    );
  }

  isInside(date: NgbDate) {
    return this.toDate && date.after(this.fromDate) && date.before(this.toDate);
  }

  isRange(date: NgbDate) {
    return (
      date.equals(this.fromDate) ||
      (this.toDate && date.equals(this.toDate)) ||
      this.isInside(date) ||
      this.isHovered(date)
    );
  }
}
