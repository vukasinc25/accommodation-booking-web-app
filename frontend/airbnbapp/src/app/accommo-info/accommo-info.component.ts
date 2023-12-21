import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { AccommodationService } from '../service/accommodation.service';
import { Accommodation } from '../model/accommodation';
import { NgbDate, NgbDateStruct } from '@ng-bootstrap/ng-bootstrap';
import { AuthService } from '../service/auth.service';
import { ReservationService } from '../service/reservation.service';
import { ResDateRange } from '../model/dateRange';
import { DisabledDateRange } from '../model/disabledDateRange';
import {
  FormBuilder,
  FormGroup,
  ValidationErrors,
  Validators,
} from '@angular/forms';

@Component({
  selector: 'app-accommo-info',
  templateUrl: './accommo-info.component.html',
  styleUrls: ['./accommo-info.component.css'],
})
export class AccommoInfoComponent implements OnInit {
  accommodationForm!: FormGroup;
  constructor(
    private fb: FormBuilder,
    private route: ActivatedRoute,
    private router: Router,
    private accommodationService: AccommodationService,
    private authService: AuthService,
    private reservationService: ReservationService
  ) {}

  role: string = '';
  username: string = '';
  startDate: NgbDate | null = null;
  endDate: NgbDate | null = null;

  id: number = 0;
  accommodation: Accommodation = {};
  dateList: ResDateRange[] = [];
  blackDateList: DisabledDateRange[] = [];
  isDataEmpty = false;

  hoveredDate: NgbDate | null = null;
  fromDate: NgbDate | null = null;
  toDate: NgbDate | null = null;

  fromDisDate: NgbDate | null = null;
  toDisDate: NgbDate | null = null;

  ngOnInit(): void {
    this.accommodationForm = this.fb.group({
      availableFrom: ['', Validators.required],
      availableUntil: ['', Validators.required],
      pricePerNight: ['', Validators.required],
      pricePerPerson: ['', Validators.required],
    });

    this.route.params.subscribe((params) => {
      this.id = params['id'];
    });
    this.role = this.authService.getRole();
    this.username = this.authService.getUsername();

    this.accommodationService.getById(this.id).subscribe({
      next: (data) => {
        this.accommodation = data;
        // console.log(this.accommodation._id);
        // console.log(data);
      },
      error: (err) => {
        console.log(err);
        this.isDataEmpty = true;
      },
    });

    this.reservationService
      .getAvailabelDatesForAccomodation(this.id)
      .subscribe({
        next: (data) => {
          // console.log(data);
          this.startDate = new NgbDate(
            new Date(data[0].startDate).getFullYear(),
            new Date(data[0].startDate).getUTCMonth() + 1,
            new Date(data[0].startDate).getUTCDate()
          );
          this.endDate = new NgbDate(
            new Date(data[0].endDate).getFullYear(),
            new Date(data[0].endDate).getUTCMonth() + 1,
            new Date(data[0].endDate).getUTCDate()
          );
          // this.startDate = new Date(data[0].startDate);
          // this.endDate = new Date(data[0].endDate);
          // console.log(this.startDate);
          // console.log(this.endDate);
        },
        error: (err) => {
          console.log(err);
          // alert(err);
        },
      });
    this.reservationService.getReservations(this.id).subscribe({
      next: (data) => {
        // console.log(data);
        if (data != null) {
          this.dateList = data as ResDateRange[];
          for (let dateRange of this.dateList) {
            let startDate = new NgbDate(
              new Date(dateRange.begin_accomodation_date!).getFullYear(),
              new Date(dateRange.begin_accomodation_date!).getUTCMonth() + 1,
              new Date(dateRange.begin_accomodation_date!).getUTCDate()
            );
            let endDate = new NgbDate(
              new Date(dateRange.end_accomodation_date!).getFullYear(),
              new Date(dateRange.end_accomodation_date!).getUTCMonth() + 1,
              new Date(dateRange.end_accomodation_date!).getUTCDate()
            );
            let blackDateRange: DisabledDateRange = { startDate, endDate };
            this.blackDateList.push(blackDateRange);
          }
          console.log(this.blackDateList);
        }
      },
      error: (err) => {
        console.log(err);
      },
    });
  }

  isDisabled = (date: NgbDate, current?: { month: number }) => {
    for (let dateRange of this.blackDateList) {
      if (
        (date.after(dateRange.startDate) && date.before(dateRange.endDate)) ||
        date.equals(dateRange.startDate) ||
        date.equals(dateRange.endDate)
      ) {
        return true;
      }
    }

    return date.after(this.fromDisDate) && date.before(this.toDisDate);
  };

  onDateSelection(date: NgbDate) {
    if (this.blackDateList.length > 0) {
      for (let blackDateRange of this.blackDateList) {
        if (!this.fromDate && !this.toDate) {
          this.fromDate = date;
        } else if (
          this.fromDate &&
          !this.toDate &&
          date.after(this.fromDate) &&
          ((this.fromDate.before(blackDateRange.startDate) &&
            date.before(blackDateRange.startDate)) ||
            (this.fromDate.after(blackDateRange.endDate) &&
              date.after(blackDateRange.endDate)))
        ) {
          this.toDate = date;
        } else if (
          this.fromDate &&
          this.toDate &&
          ((this.fromDate.before(blackDateRange.startDate) &&
            date.before(blackDateRange.startDate)) ||
            (this.fromDate.after(blackDateRange.endDate) &&
              date.after(blackDateRange.endDate)))
        ) {
          continue;
        } else {
          this.toDate = null;
          this.fromDate = date;
        }
      }
    } else {
      if (!this.fromDate && !this.toDate) {
        this.fromDate = date;
      } else if (this.fromDate && !this.toDate && date.after(this.fromDate)) {
        this.toDate = date;
      } else {
        this.toDate = null;
        this.fromDate = date;
      }
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

  reserveDates() {
    this.reservationService
      .createReservation(this.accommodation._id!, this.fromDate!, this.toDate!)
      .subscribe({
        next: (data) => {
          console.log(data);
        },
        error: (err) => {
          console.log(err);
        },
      });
  }
  onSubmit() {
    console.log(this.accommodation._id, this.accommodationForm.value);
    this.reservationService
      .createReservationDatesForAccomodation(
        this.accommodation._id,
        this.accommodationForm.value
      )
      .subscribe({
        next: (data) => {
          console.log('Reservation in succesfuly created');
          alert('Reservation is successfully.');
          this.router.navigate(['accommodations/myAccommodations']);
        },
        error: (err) => {
          console.log(err.error.message);
          alert(err.error.message);
        },
      });
  }
}
